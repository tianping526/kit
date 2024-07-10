package pipeline

import (
	"context"
	"reflect"
	"strconv"
	"sync"
	"testing"
	"time"
)

func TestPipeline(t *testing.T) {
	conf := &Config{
		MaxSize:  3,
		Interval: time.Millisecond * 20,
		Buffer:   3,
		Worker:   10,
	}
	type recv struct {
		mirror bool
		ch     int
		values map[string][]interface{}
	}
	var runs []recv
	mu := sync.Mutex{}
	do := func(mirror bool, ch int, values map[string][]interface{}) {
		mu.Lock()
		defer mu.Unlock()
		runs = append(runs, recv{
			mirror: mirror,
			values: values,
			ch:     ch,
		})
	}
	split := func(s string) int {
		n, _ := strconv.Atoi(s)
		return n
	}
	p := NewPipeline(conf)
	p.Do = do
	p.Split = split
	p.Start()
	_ = p.Add("1", 1)
	_ = p.Add("1", 2)
	_ = p.Add("11", 3)
	_ = p.Add("2", 3)
	time.Sleep(time.Millisecond * 60)
	_ = p.MirrorAdd("2", 3)
	time.Sleep(time.Millisecond * 60)
	_ = p.SyncMirrorAdd(context.Background(), "5", 5)
	time.Sleep(time.Millisecond * 60)
	_ = p.Close()
	expt := []recv{
		{
			mirror: false,
			ch:     1,
			values: map[string][]interface{}{
				"1":  {1, 2},
				"11": {3},
			},
		},
		{
			mirror: false,
			ch:     2,
			values: map[string][]interface{}{
				"2": {3},
			},
		},
		{
			mirror: true,
			ch:     2,
			values: map[string][]interface{}{
				"2": {3},
			},
		},
		{
			mirror: true,
			ch:     5,
			values: map[string][]interface{}{
				"5": {5},
			},
		},
	}
	mu.Lock()
	defer mu.Unlock()
	if !reflect.DeepEqual(runs, expt) {
		t.Errorf("expect get %+v,\n got: %+v", expt, runs)
	}
}

func TestPipelineSmooth(t *testing.T) {
	conf := &Config{
		MaxSize:  100,
		Interval: time.Second,
		Buffer:   100,
		Worker:   10,
	}
	type result struct {
		index int
		ts    time.Time
	}
	var results []result
	mu := sync.Mutex{}
	wg := sync.WaitGroup{}
	wg.Add(10)
	do := func(_ bool, index int, _ map[string][]interface{}) {
		defer wg.Done()
		mu.Lock()
		defer mu.Unlock()
		results = append(results, result{
			index: index,
			ts:    time.Now(),
		})
	}
	split := func(s string) int {
		n, _ := strconv.Atoi(s)
		return n
	}
	p := NewPipeline(conf)
	p.Do = do
	p.Split = split
	p.Start()
	for i := 0; i < 10; i++ {
		_ = p.Add(strconv.Itoa(i), 1)
	}
	wg.Wait()
	mu.Lock()
	defer mu.Unlock()
	if len(results) != conf.Worker {
		t.Errorf("expect results equal worker")
	}
	for i, r := range results {
		if i > 0 {
			if r.ts.Sub(results[i-1].ts) < time.Millisecond*20 {
				t.Errorf("expect runs be smooth")
			}
		}
	}
}
