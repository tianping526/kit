package pipeline

import (
	"context"
	"errors"
	"strconv"
	"sync"
	"time"

	"github.com/go-kratos/kratos/v2/metrics"
)

// ErrFull channel full error
var ErrFull = errors.New("channel full")

type message struct {
	key   string
	value interface{}
}

// Pipeline pipeline struct
type Pipeline struct {
	Do          func(mirror bool, index int, values map[string][]interface{})
	Split       func(key string) int
	chans       []chan *message
	mirrorChans []chan *message
	config      *Config
	wait        sync.WaitGroup
	name        string
}

// Config Pipeline config
type Config struct {
	// MaxSize merge size
	MaxSize int
	// Interval merge interval
	Interval time.Duration
	// Buffer channel size
	Buffer int
	// Worker channel number
	Worker int
	// Name use for metrics
	Name string
	// ChanLen use for metric channel length
	ChanLen metrics.Gauge
	// ChanLen use for metric item count
	Count metrics.Counter
}

func (c *Config) fix() {
	if c.MaxSize <= 0 {
		c.MaxSize = 1000
	}
	if c.Interval <= 0 {
		c.Interval = time.Second
	}
	if c.Buffer <= 0 {
		c.Buffer = 1000
	}
	if c.Worker <= 0 {
		c.Worker = 10
	}
	if c.Name == "" {
		c.Name = "anonymous"
	}
}

// NewPipeline new pipeline
func NewPipeline(config *Config) (res *Pipeline) {
	if config == nil {
		config = &Config{}
	}
	config.fix()
	res = &Pipeline{
		chans:       make([]chan *message, config.Worker),
		mirrorChans: make([]chan *message, config.Worker),
		config:      config,
		name:        config.Name,
	}
	for i := 0; i < config.Worker; i++ {
		res.chans[i] = make(chan *message, config.Buffer)
		res.mirrorChans[i] = make(chan *message, config.Buffer)
	}
	return
}

// Start all mergeProc
func (p *Pipeline) Start() {
	if p.Do == nil {
		panic("pipeline: do func is nil")
	}
	if p.Split == nil {
		panic("pipeline: split func is nil")
	}
	var mirror bool
	p.wait.Add(len(p.chans) + len(p.mirrorChans))
	for i, ch := range p.chans {
		go p.mergeProc(mirror, i, ch)
	}
	mirror = true
	for i, ch := range p.mirrorChans {
		go p.mergeProc(mirror, i, ch)
	}
}

// SyncAdd sync add a value to channel, channel shard in split method
func (p *Pipeline) SyncAdd(c context.Context, key string, value interface{}) (err error) {
	ch, msg := p.add(false, key, value)
	select {
	case ch <- msg:
	case <-c.Done():
		err = c.Err()
	}
	return
}

// Add async add a value to channel, channel shard in split method
func (p *Pipeline) Add(key string, value interface{}) (err error) {
	ch, msg := p.add(false, key, value)
	select {
	case ch <- msg:
	default:
		err = ErrFull
	}
	return
}

// SyncMirrorAdd sync add a value to mirror channel, channel shard in split method
func (p *Pipeline) SyncMirrorAdd(c context.Context, key string, value interface{}) (err error) {
	ch, msg := p.add(true, key, value)
	select {
	case ch <- msg:
	case <-c.Done():
		err = c.Err()
	}
	return
}

// MirrorAdd async add a value to mirror channel, channel shard in split method
func (p *Pipeline) MirrorAdd(key string, value interface{}) (err error) {
	ch, msg := p.add(true, key, value)
	select {
	case ch <- msg:
	default:
		err = ErrFull
	}
	return
}

func (p *Pipeline) add(mirror bool, key string, value interface{}) (ch chan *message, m *message) {
	shard := p.Split(key) % p.config.Worker
	if mirror {
		ch = p.mirrorChans[shard]
	} else {
		ch = p.chans[shard]
	}
	m = &message{key: key, value: value}
	return
}

// Close all goroutine
func (p *Pipeline) Close() (err error) {
	for _, ch := range p.chans {
		ch <- nil
	}
	for _, ch := range p.mirrorChans {
		ch <- nil
	}
	p.wait.Wait()
	return
}

func (p *Pipeline) mergeProc(mirror bool, index int, ch <-chan *message) {
	defer p.wait.Done()
	var (
		m        *message
		vals     = make(map[string][]interface{}, p.config.MaxSize)
		closed   bool
		count    int
		interval = p.config.Interval
		timeout  = false
	)
	if index > 0 {
		interval = time.Duration(int64(index) * (int64(p.config.Interval) / int64(p.config.Worker)))
	}
	timer := time.NewTimer(interval)
	defer timer.Stop()
	for {
		select {
		case m = <-ch:
			if m == nil {
				closed = true
				break
			}
			count++
			vals[m.key] = append(vals[m.key], m.value)
			if count >= p.config.MaxSize {
				break
			}
			continue
		case <-timer.C:
			timeout = true
		}
		name := p.name
		process := count
		if len(vals) > 0 {
			if mirror {
				name = "mirror_" + name
			}
			p.Do(mirror, index, vals)
			vals = make(map[string][]interface{}, p.config.MaxSize)
			count = 0
		}
		if p.config.ChanLen != nil {
			p.config.ChanLen.With(name, strconv.Itoa(index)).Set(float64(len(ch)))
		}
		if p.config.Count != nil {
			p.config.Count.With(name, strconv.Itoa(index)).Add(float64(process))
		}
		if closed {
			return
		}
		if !timer.Stop() && !timeout {
			<-timer.C
			timeout = false
		}
		timer.Reset(p.config.Interval)
	}
}
