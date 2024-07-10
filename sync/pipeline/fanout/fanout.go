package fanout

import (
	"context"
	"errors"
	"fmt"
	"runtime"
	"sync"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/metadata"
	"github.com/go-kratos/kratos/v2/metrics"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

var (
	// ErrFull chan full.
	ErrFull = errors.New("fanout: chan full")
	tracer  = otel.Tracer("github.com/tianping526/kit/sync/pipeline/fanout")
)

type options struct {
	worker        int
	buffer        int
	metricChanLen metrics.Gauge
	metricCount   metrics.Counter
	logger        *log.Helper
}

// Option fanout option
type Option func(*options)

// WithWorker specifies the worker of fanout
func WithWorker(n int) Option {
	if n <= 0 {
		panic("fanout: worker should > 0")
	}
	return func(o *options) {
		o.worker = n
	}
}

// WithBuffer specifies the buffer of fanout
func WithBuffer(n int) Option {
	if n <= 0 {
		panic("fanout: buffer should > 0")
	}
	return func(o *options) {
		o.buffer = n
	}
}

// WithMetricChanLen specifies the metric of fanout channel length
func WithMetricChanLen(m metrics.Gauge) Option {
	return func(o *options) {
		o.metricChanLen = m
	}
}

// WithMetricCount specifies the metric of fanout item count
func WithMetricCount(m metrics.Counter) Option {
	return func(o *options) {
		o.metricCount = m
	}
}

// WithLogger specifies the logger of fanout
func WithLogger(l log.Logger) Option {
	return func(o *options) {
		o.logger = log.NewHelper(log.With(l, "module", "fanout", "caller", log.DefaultCaller))
	}
}

type item struct {
	f   func(c context.Context)
	ctx context.Context
}

// Fanout async consume data from chan.
type Fanout struct {
	name    string
	ch      chan item
	options *options
	waiter  sync.WaitGroup

	ctx    context.Context
	cancel func()
}

// New a fanout struct.
func New(name string, opts ...Option) *Fanout {
	if name == "" {
		name = "anonymous"
	}
	o := &options{
		worker: 1,
		buffer: 1024,
		logger: log.NewHelper(log.With(log.DefaultLogger, "module", "fanout", "caller", log.DefaultCaller)),
	}
	for _, op := range opts {
		op(o)
	}
	c := &Fanout{
		ch:      make(chan item, o.buffer),
		name:    name,
		options: o,
	}
	c.ctx, c.cancel = context.WithCancel(context.Background())
	c.waiter.Add(o.worker)
	for i := 0; i < o.worker; i++ {
		go c.proc()
	}
	return c
}

func (c *Fanout) proc() {
	defer c.waiter.Done()
	for {
		select {
		case t := <-c.ch:
			wrapFunc(t.f, c.options.logger)(t.ctx)
			if c.options.metricChanLen != nil {
				c.options.metricChanLen.With(c.name).Set(float64(len(c.ch)))
			}
			if c.options.metricCount != nil {
				c.options.metricCount.With(c.name).Inc()
			}
		case <-c.ctx.Done():
			return
		}
	}
}

func wrapFunc(f func(c context.Context), logger *log.Helper) (res func(context.Context)) {
	res = func(ctx context.Context) {
		span := trace.SpanFromContext(ctx)
		defer span.End()
		defer func() {
			if r := recover(); r != nil {
				buf := make([]byte, 64<<10)
				buf = buf[:runtime.Stack(buf, false)]
				logger.WithContext(ctx).Errorf("err: %s, stack: %s", r, buf)
				span.SetStatus(codes.Error, fmt.Sprintf("%s", r))
			}
		}()
		f(ctx)
	}
	return
}

// Do save a callback func.
func (c *Fanout) Do(ctx context.Context, f func(ctx context.Context)) (err error) {
	if f == nil || c.ctx.Err() != nil {
		return c.ctx.Err()
	}
	newCtx := context.Background()
	md, ok := metadata.FromServerContext(ctx)
	if ok {
		nmd := md.Clone()
		newCtx = metadata.NewServerContext(newCtx, nmd)
	}
	_, span := tracer.Start(ctx, "fanout:Do", trace.WithSpanKind(trace.SpanKindInternal))
	sc := trace.ContextWithSpan(newCtx, span)
	select {
	case c.ch <- item{f: f, ctx: sc}:
	default:
		err = ErrFull
		span.SetStatus(codes.Error, err.Error())
		span.End()
	}
	if c.options.metricChanLen != nil {
		c.options.metricChanLen.With(c.name).Set(float64(len(c.ch)))
	}
	return
}

// Close fanout
func (c *Fanout) Close() error {
	if err := c.ctx.Err(); err != nil {
		return err
	}
	c.cancel()
	c.waiter.Wait()
	return nil
}
