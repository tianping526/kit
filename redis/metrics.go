package redis

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/go-kratos/kratos/v2/metrics"
	"github.com/redis/go-redis/extra/rediscmd/v9"
	"github.com/redis/go-redis/v9"
)

var _ redis.Hook = (*MetricHook)(nil)

// Option is metrics option.
type Option func(*options)

// WithRequestsDuration with requests duration(s).
func WithRequestsDuration(c metrics.Observer) Option {
	return func(o *options) {
		o.requestsDuration = c
	}
}

// WithAddr with db Addr.
func WithAddr(a string) Option {
	return func(o *options) {
		o.Addr = a
	}
}

func NewMetricHook(opts ...Option) *MetricHook {
	op := options{}
	for _, o := range opts {
		o(&op)
	}

	return &MetricHook{op: op}
}

type options struct {
	// histogram: db_client_requests_duration_sec_bucket{"name", "addr", "res"}
	requestsDuration metrics.Observer
	Addr             string
}

type MetricHook struct {
	op options
}

func (mh *MetricHook) DialHook(next redis.DialHook) redis.DialHook {
	return func(ctx context.Context, network, addr string) (net.Conn, error) {
		return next(ctx, network, addr)
	}
}

func (mh *MetricHook) ProcessHook(next redis.ProcessHook) redis.ProcessHook {
	return func(ctx context.Context, cmd redis.Cmder) error {
		var err error
		if mh.op.requestsDuration != nil {
			startTime := time.Now()
			defer func() {
				res := "ok"
				if err != nil && err != redis.Nil {
					res = fmt.Sprintf("%T", err)
				}
				mh.op.requestsDuration.
					With(cmd.FullName(), mh.op.Addr, res).
					Observe(time.Since(startTime).Seconds())
			}()
		}
		err = next(ctx, cmd)
		return err
	}
}

func (mh *MetricHook) ProcessPipelineHook(next redis.ProcessPipelineHook) redis.ProcessPipelineHook {
	return func(ctx context.Context, cmds []redis.Cmder) error {
		var err error
		if mh.op.requestsDuration != nil {
			startTime := time.Now()
			defer func() {
				summary, _ := rediscmd.CmdsString(cmds)
				res := "ok"
				if err != nil && err != redis.Nil {
					res = fmt.Sprintf("%T", err)
				}
				mh.op.requestsDuration.
					With(fmt.Sprintf("pipeline%s", summary), mh.op.Addr, res).
					Observe(time.Since(startTime).Seconds())
			}()
		}
		err = next(ctx, cmds)
		return err
	}
}
