package server

const serviceTemplate = `
{{- /* delete empty line */ -}}
package service

import (
	{{- if .UseContext }}
	"context"
	{{- end }}
	{{- if .UseIO }}
	"io"
	{{- end }}

	{{ $.Version }} "{{ .Package }}"
	"{{ .ServicePath }}/internal/biz"
	"github.com/go-kratos/kratos/v2/log"
	{{- if .GoogleEmpty }}
	"google.golang.org/protobuf/types/known/emptypb"
	{{- end }}
)

type {{ .Service }} struct {
	{{ $.Version }}.Unimplemented{{ .Service }}Server

	uc  *biz.{{ .ServiceName }}UseCase
	log *log.Helper
}

func New{{ .Service }}(uc *biz.{{ .ServiceName }}UseCase, logger log.Logger) *{{ .Service }} {
	return &{{ .Service }}{
		log: log.NewHelper(log.With(
			logger,
			"module", "service/{{ .Service }}",
			"caller", log.DefaultCaller,
		)),
		uc: uc,
	}
}

{{- $s1 := "google.protobuf.Empty" }}
{{- range .Methods }}
{{- if eq .Type 1 }}

func (s *{{ .Service }}) {{ .Name }}(
	ctx context.Context,
	req {{ if eq .Request $s1 }}*emptypb.Empty{{ else }}*{{ $.Version }}.{{ .Request }}{{ end }},
) ({{ if eq .Reply $s1 }}*emptypb.Empty{{ else }}*{{ $.Version }}.{{ .Reply }}{{ end }}, error) {
{{ .ReqCopy }}
	{{ .RepDoName }}, err := s.uc.{{ .Name }}(ctx, {{ .Request | dtoCovertDoName | toInternalName }})
	if err != nil {
		return nil, err
	}
{{ .RepCopy }}
	return {{ .Reply | toInternalName }}, nil
}
{{- else if eq .Type 2 }}

func (s *{{ .Service }}) {{ .Name }}(conn {{ $.Version }}.{{ .Service }}_{{ .Name }}Server) error {
	for {
		req, err := conn.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
		
		err = conn.Send(&{{ $.Version }}.{{ .Reply }}{})
		if err != nil {
			return err
		}
	}
}
{{- else if eq .Type 3 }}

func (s *{{ .Service }}) {{ .Name }}(conn {{ $.Version }}.{{ .Service }}_{{ .Name }}Server) error {
	for {
		req, err := conn.Recv()
		if err == io.EOF {
			return conn.SendAndClose(&{{ $.Version }}.{{ .Reply }}{})
		}
		if err != nil {
			return err
		}
	}
}
{{- else if eq .Type 4 }}

func (s *{{ .Service }}) {{ .Name }}(req {{ if eq .Request $s1 }}*emptypb.Empty
{{ else }}*{{ $.Version }}.{{ .Request }}{{ end }}, conn {{ $.Version }}.{{ .Service }}_{{ .Name }}Server) error {
	for {
		err := conn.Send(&{{ $.Version }}.{{ .Reply }}{})
		if err != nil {
			return err
		}
	}
}
{{- end }}
{{- end }}
`

const serviceReadmeTemplate = `
{{- /* delete empty line */ -}}
# Service
`

const serviceServiceTemplate = `
{{- /* delete empty line */ -}}
package service

import (
	"github.com/google/wire"
)

// ProviderSet is service providers.
var ProviderSet = wire.NewSet(
{{- range . }}
	New{{ .Service }},
{{ end -}}
)
`

const bizTemplate = `
{{- /* delete empty line */ -}}
package biz

import (
	{{- if .UseContext }}
	"context"
	{{- end }}
	{{- if .UseIO }}
	"io"
	{{- end }}

	"github.com/go-kratos/kratos/v2/log"
	{{- if .GoogleEmpty }}
	"google.golang.org/protobuf/types/known/emptypb"
	{{- end }}
	{{- range .Imports }}
	"{{ . }}"
	{{- end }}
)

{{- if ne .DO "" -}}
{{ .DO }}
{{- end }}

{{ $s1 := "google.protobuf.Empty" -}}
type {{ .ServiceName }}Repo interface {
{{ range .Methods }}	{{ .Name }}(ctx context.Context, param {{ if eq .Request $s1 -}}
*emptypb.Empty{{ else }}*{{ .Request | dtoCovertDoName }}{{ end }}) ({{ if eq .Reply $s1 -}}
*emptypb.Empty{{ else }}*{{ .Reply | dtoCovertDoName }}{{ end }}, error)
{{ end -}}
}

type {{ .ServiceName }}UseCase struct {
	repo {{ .ServiceName }}Repo
	log  *log.Helper
}

func New{{ .ServiceName }}UseCase(
	repo {{ .ServiceName }}Repo,
	logger log.Logger,
) *{{ .ServiceName }}UseCase {
	return &{{ .ServiceName }}UseCase{
		repo: repo,
		log: log.NewHelper(log.With(
			logger,
			"module", "biz/{{ .ServiceName }}",
			"caller", log.DefaultCaller,
		)),
	}
}

{{- range .Methods }}

func (uc *{{ $.ServiceName }}UseCase) {{ .Name }}(
	ctx context.Context,
	param {{ if eq .Request $s1 }}*emptypb.Empty{{ else }}*{{ .Request | dtoCovertDoName }}{{ end }},
) ({{ if eq .Reply $s1 }}*emptypb.Empty{{ else }}*{{ .Reply | dtoCovertDoName }}{{ end }}, error) {
	do, err := uc.repo.{{ .Name }}(ctx, param)
	if err != nil {
		return nil, err
	}
	return do, nil
}
{{- end }}
`

const bizReadmeTemplate = `
{{- /* delete empty line */ -}}
# Biz
`

const bizBizTemplate = `
{{- /* delete empty line */ -}}
package biz

import (
	"github.com/google/wire"
)

// ProviderSet is service providers.
var ProviderSet = wire.NewSet(
{{- range . }}
	New{{ .ServiceName }}UseCase,
{{ end -}}
)
`

const dataTemplate = `
{{- /* delete empty line */ -}}
package data

import (
	{{- if .UseContext }}
	"context"
	{{- end }}
	{{- if .UseIO }}
	"io"
	{{- end }}

	"{{ .ServicePath }}/internal/biz"
	"github.com/go-kratos/kratos/v2/log"
	{{- if .GoogleEmpty }}
	"google.golang.org/protobuf/types/known/emptypb"
	{{- end }}
)

var _ biz.{{ .ServiceName }}Repo = (*{{ .ServiceInternalName }}Repo)(nil)

{{ $s1 := "google.protobuf.Empty" -}}
type {{ .ServiceInternalName }}Repo struct {
	data *Data
	log  *log.Helper
}

func New{{ .ServiceName }}Repo(
	data *Data,
	logger log.Logger,
) biz.{{ .ServiceName }}Repo {
	return &{{ .ServiceInternalName }}Repo{
		data: data,
		log: log.NewHelper(log.With(
			logger,
			"module", "data/{{ .ServiceName }}",
			"caller", log.DefaultCaller,
		)),
	}
}

{{- range .Methods }}

func (repo *{{ $.ServiceInternalName }}Repo) {{ .Name }}(
	ctx context.Context,
	param {{ if eq .Request $s1 }}*emptypb.Empty{{ else }}*biz.{{ .Request | dtoCovertDoName }}{{ end }},
) ({{ if eq .Reply $s1 }}*emptypb.Empty{{ else }}*biz.{{ .Reply | dtoCovertDoName }}{{ end }}, error) {
	return {{ if eq .Reply $s1 }}&emptypb.Empty{{ else }}&biz.{{ .Reply | dtoCovertDoName }}{{ end }}{}, nil
}
{{- end }}
`

const dataReadmeTemplate = `
{{- /* delete empty line */ -}}
# Data
`

const dataDataTemplate = `
{{- /* delete empty line */ -}}
package data

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/url"
	"regexp"
	"strings"
	"time"

	"{{ (index . 0).ServicePath }}/internal/conf"
	"{{ (index . 0).ServicePath }}/internal/data/ent"
	"{{ (index . 0).ServicePath }}/internal/data/ent/hook"
	"{{ (index . 0).ServicePath }}/internal/data/ent/migrate"
	"entgo.io/ent/dialect"
	entSql "entgo.io/ent/dialect/sql"
	"github.com/go-kratos/kratos/contrib/config/apollo/v2"
	kz "github.com/go-kratos/kratos/contrib/log/zap/v2"
	prom "github.com/go-kratos/kratos/contrib/metrics/prometheus/v2"
	"github.com/go-kratos/kratos/contrib/registry/consul/v2"
	"github.com/go-kratos/kratos/v2/config"
	"github.com/go-kratos/kratos/v2/config/file"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/metrics"
	"github.com/go-kratos/kratos/v2/middleware/tracing"
	"github.com/go-kratos/kratos/v2/registry"
	"github.com/google/wire"
	consulAPI "github.com/hashicorp/consul/api"
	"github.com/natefinch/lumberjack"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/redis/go-redis/extra/redisotel/v9"
	"github.com/redis/go-redis/v9"
	"github.com/signalfx/splunk-otel-go/instrumentation/database/sql/splunksql"
	mr "github.com/tianping526/kit/redis"
	"github.com/tianping526/kit/sync/pipeline/fanout"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/sdk/resource"
	traceSDK "go.opentelemetry.io/otel/sdk/trace"
	semConv "go.opentelemetry.io/otel/semconv/v1.7.0"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	// init db driver
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/jackc/pgx/v4/stdlib"
)

// ProviderSet is data providers.
var ProviderSet = wire.NewSet(
	NewLogger,
	NewData,
	NewConfig,
	NewConfigBootstrap,
	NewRegistrar,
	NewMetric,
	NewTracerProvider,
	NewEntClient,
	NewRedisCmd,
	NewCacheFanout,
{{- range . }}
	New{{ .ServiceName }}Repo,
{{ end -}}
)

// Data .
type Data struct {
	log *log.Helper
	cf  *fanout.Fanout
	m   *Metric

	db *ent.Client
	rc redis.Cmdable
}

type Metric struct {
	CacheHits        metrics.Counter
	CacheMisses      metrics.Counter
	CacheDurationSec metrics.Observer
	CodeTotal        metrics.Counter
	DurationSec      metrics.Observer
	DbDurationSec    metrics.Observer
	CacheChanLen     metrics.Gauge
	CacheCount       metrics.Counter
	// PipelineChanLen metrics.Gauge
	// PipelineCount   metrics.Counter
}

// NewData .
func NewData(
	logger log.Logger,
	db *ent.Client,
	rcd redis.Cmdable,
	cf *fanout.Fanout,
	m *Metric,
) (*Data, error) {
	l := log.NewHelper(log.With(
		logger,
		"module", "data/Data",
		"caller", log.DefaultCaller,
	))
	rc, _ := rcd.(*redis.Client)
	return &Data{
		log: l,
		db:  db,
		rc:  rc,
		cf:  cf,
		m:   m,
	}, nil
}

func NewMetric() *Metric {
	cacheHits := prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "cache",
		Subsystem: "redis",
		Name:      "hits_total",
		Help:      "redis hits total.",
	}, []string{"name"})
	cacheMisses := prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "cache",
		Subsystem: "redis",
		Name:      "misses_total",
		Help:      "redis misses total.",
	}, []string{"name"})
	cacheDurationSec := prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: "cache_client",
		Subsystem: "requests",
		Name:      "duration_sec",
		Help:      "Cache requests duration(sec).",
		Buckets:   []float64{0.005, 0.01, 0.025, 0.05, 0.1, 0.250, 0.5, 1},
	}, []string{"name", "addr", "res"})
	codeTotal := prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "server",
		Subsystem: "requests",
		Name:      "code_total",
		Help:      "The total number of processed requests",
	}, []string{"kind", "operation", "code", "reason"})
	durationSec := prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: "server",
		Subsystem: "requests",
		Name:      "duration_sec",
		Help:      "Server requests duration(sec).",
		Buckets:   []float64{0.005, 0.01, 0.025, 0.05, 0.1, 0.250, 0.5, 1},
	}, []string{"kind", "operation"})
	dbDurationSec := prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: "db_client",
		Subsystem: "requests",
		Name:      "duration_sec",
		Help:      "DB requests duration(sec).",
		Buckets:   []float64{0.005, 0.01, 0.025, 0.05, 0.1, 0.250, 0.5, 1},
	}, []string{"name", "addr", "command", "res"})
	cacheChanLen := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "sync",
		Subsystem: "pipeline_fanout",
		Name:      "chan_len",
		Help:      "sync pipeline fanout current channel size.",
	}, []string{"name"})
	cacheCount := prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "sync",
		Subsystem: "pipeline_fanout",
		Name:      "process_count",
		Help:      "process count",
	}, []string{"name"})
	//pipelineChanLen := prometheus.NewGaugeVec(prometheus.GaugeOpts{
	//	Namespace: "sync",
	//	Subsystem: "pipeline",
	//	Name:      "chan_len",
	//	Help:      "channel length",
	//}, []string{"name", "chan"})
	//pipelineCount := prometheus.NewCounterVec(prometheus.CounterOpts{
	//	Namespace: "sync",
	//	Subsystem: "pipeline",
	//	Name:      "process_count",
	//	Help:      "process count",
	//}, []string{"name", "chan"})

	prometheus.MustRegister(
		cacheHits, cacheMisses, codeTotal,
		durationSec, cacheChanLen, cacheCount,
		dbDurationSec, cacheDurationSec,
		// pipelineChanLen, pipelineCount,
	)
	return &Metric{
		CacheHits:        prom.NewCounter(cacheHits),
		CacheMisses:      prom.NewCounter(cacheMisses),
		CacheDurationSec: prom.NewHistogram(cacheDurationSec),
		CodeTotal:        prom.NewCounter(codeTotal),
		DurationSec:      prom.NewHistogram(durationSec),
		DbDurationSec:    prom.NewHistogram(dbDurationSec),
		CacheChanLen:     prom.NewGauge(cacheChanLen),
		CacheCount:       prom.NewCounter(cacheCount),
		// PipelineChanLen: prom.NewGauge(pipelineChanLen),
		// PipelineCount:   prom.NewCounter(pipelineCount),
	}
}

func NewLogger(ai *conf.AppInfo, cfg *conf.Bootstrap) (log.Logger, func(), error) {
	level := conf.Log_INFO
	encoding := conf.Log_JSON
	sampling := &zap.SamplingConfig{
		Initial:    100,
		Thereafter: 100,
	}
	outputPaths := []*conf.Log_Output{ {{- /* place */ -}}{Path: "stderr"}}

	if cfg.Log != nil {
		level = cfg.Log.Level
		encoding = cfg.Log.Encoding
		if cfg.Log.Sampling != nil {
			sampling = &zap.SamplingConfig{
				Initial:    int(cfg.Log.Sampling.Initial),
				Thereafter: int(cfg.Log.Sampling.Thereafter),
			}
		}
		if len(cfg.Log.OutputPaths) > 0 {
			outputPaths = cfg.Log.OutputPaths
		}
	}

	// encoder
	var encoder zapcore.Encoder
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.TimeKey = zapcore.OmitKey
	encoderConfig.NameKey = zapcore.OmitKey
	encoderConfig.CallerKey = zapcore.OmitKey
	encoderConfig.MessageKey = zapcore.OmitKey
	if encoding == conf.Log_CONSOLE {
		encoder = zapcore.NewConsoleEncoder(encoderConfig)
	} else {
		encoder = zapcore.NewJSONEncoder(encoderConfig)
	}

	// sinks
	var sink zapcore.WriteSyncer
	closes := make([]func(), 0, len(outputPaths))
	paths := make([]string, 0, len(outputPaths))
	syncer := make([]zapcore.WriteSyncer, 0, len(outputPaths))
	for _, out := range outputPaths {
		if out.Rotate == nil {
			paths = append(paths, out.Path)
			continue
		}

		lg := &lumberjack.Logger{
			Filename:   out.Path,
			MaxSize:    int(out.Rotate.MaxSize),
			MaxAge:     int(out.Rotate.MaxAge),
			MaxBackups: int(out.Rotate.MaxBackups),
			Compress:   out.Rotate.Compress,
		}

		syncer = append(syncer, zapcore.AddSync(lg))
		closes = append(closes, func() {
			err := lg.Close()
			if err != nil {
				fmt.Printf("close lumberjack logger(%s) error(%s))", out.Path, err)
			}
		})
	}
	if len(paths) > 0 {
		writer, mc, err := zap.Open(paths...)
		if err != nil {
			for _, c := range closes {
				c()
			}
			return nil, nil, err
		}
		closes = append(closes, mc)
		syncer = append(syncer, writer)
	}
	sink = zap.CombineWriteSyncers(syncer...)

	zl := zap.New(
		zapcore.NewCore(encoder, sink, zap.NewAtomicLevelAt(zapcore.Level(level-1))),
		zap.AddStacktrace(zapcore.ErrorLevel),
		zap.WrapCore(func(core zapcore.Core) zapcore.Core {
			return zapcore.NewSamplerWithOptions(
				core,
				time.Second,
				sampling.Initial,
				sampling.Thereafter,
			)
		}),
	)

	logger := log.With(
		kz.NewLogger(zl),
		"ts", log.DefaultTimestamp,
		"service.id", ai.Id,
		"service.name", ai.Name,
		"service.version", ai.Version,
		"trace_id", tracing.TraceID(),
		"span_id", tracing.SpanID(),
	)
	return logger, func() {
		err := zl.Sync()
		if err != nil {
			fmt.Printf("sync logger error(%s)", err)
		}
		for _, c := range closes {
			c()
		}
	}, nil
}

func NewConfig(ai *conf.AppInfo) (config.Config, func(), error) {
	fc := config.New(
		config.WithSource(
			file.NewSource(ai.FlagConf),
		),
	)
	if err := fc.Load(); err != nil {
		return nil, nil, err
	}

	_, err := fc.Value("apollo").Map()
	if errors.Is(err, config.ErrNotFound) {
		return fc, func() {
			err := fc.Close()
			if err != nil {
				fmt.Printf("close file config(%s) error(%s))", ai.FlagConf, err)
			}
		}, nil
	}

	var fco conf.ApolloConfig
	if err := fc.Scan(&fco); err != nil {
		return nil, nil, err
	}
	if err := fc.Close(); err != nil {
		return nil, nil, err
	}

	namePart := strings.Split(ai.Name, ".")
	var build strings.Builder
	build.WriteString(namePart[2])
	build.WriteString(".yaml")
	af := config.New(
		config.WithSource(
			apollo.NewSource(
				apollo.WithAppID(strings.Join(namePart[:2], ".")),
				apollo.WithCluster(fco.Apollo.Cluster),
				apollo.WithEndpoint(fco.Apollo.Endpoint),
				apollo.WithNamespace(build.String()),
				apollo.WithSecret(fco.Apollo.Secret),
			),
		),
	)
	build.Reset()
	if err := af.Load(); err != nil {
		return nil, nil, err
	}

	return af, func() {
		err := af.Close()
		if err != nil {
			fmt.Printf("close apollo config(%s) error(%s))", ai.Name, err)
		}
	}, nil
}

func NewConfigBootstrap(c config.Config) (*conf.Bootstrap, error) {
	var bc conf.Bootstrap
	if err := c.Value("bootstrap").Scan(&bc); err != nil {
		return nil, err
	}

	return &bc, nil
}

func NewDiscovery(conf *conf.Bootstrap) (registry.Discovery, error) {
	if conf.Registry == nil {
		return nil, nil
	}
	c := consulAPI.DefaultConfig()
	c.Address = conf.Registry.Consul.Address
	c.Scheme = conf.Registry.Consul.Scheme
	cli, err := consulAPI.NewClient(c)
	if err != nil {
		return nil, err
	}
	r := consul.New(cli)
	return r, nil
}

func NewRegistrar(conf *conf.Bootstrap) (registry.Registrar, error) {
	if conf.Registry == nil {
		return nil, nil
	}
	c := consulAPI.DefaultConfig()
	c.Address = conf.Registry.Consul.Address
	c.Scheme = conf.Registry.Consul.Scheme
	cli, err := consulAPI.NewClient(c)
	if err != nil {
		return nil, err
	}
	r := consul.New(cli)
	return r, nil
}

func NewTracerProvider(ai *conf.AppInfo, conf *conf.Bootstrap) (trace.TracerProvider, func(), error) {
	if conf.Trace == nil {
		return trace.NewNoopTracerProvider(), func() {}, nil
	}
	exp, err := jaeger.New(
		jaeger.WithCollectorEndpoint(
			jaeger.WithEndpoint(conf.Trace.Endpoint),
		),
	)
	if err != nil {
		return nil, nil, err
	}
	tp := traceSDK.NewTracerProvider(
		traceSDK.WithBatcher(exp),
		traceSDK.WithResource(resource.NewSchemaless(
			semConv.ServiceNameKey.String(ai.Name),
		)),
		traceSDK.WithSampler(traceSDK.ParentBased(traceSDK.TraceIDRatioBased(1.0))),
	)
	otel.SetTracerProvider(tp)
	return tp, func() {
		err := tp.Shutdown(context.Background())
		if err != nil {
			fmt.Printf("close trace provider(%s) error(%s))", conf.Trace.Endpoint, err)
		}
	}, nil
}

type queryMetricDriver struct {
	tableNameRe *regexp.Regexp
	durationSec metrics.Observer
	addr        string
	*entSql.Driver
}

func (qmd queryMetricDriver) Query(ctx context.Context, query string, args, v interface{}) (err error) {
	res := qmd.tableNameRe.FindStringSubmatch(query)
	tableName := ""
	if len(res) > 1 {
		tableName = res[1]
	}
	if qmd.durationSec != nil {
		startTime := time.Now()
		defer func() {
			result := "ok"
			if err != nil {
				result = fmt.Sprintf("%T", err)
			}
			qmd.durationSec.
				With(tableName, qmd.addr, "query", result).
				Observe(time.Since(startTime).Seconds())
		}()
	}

	err = qmd.Driver.Query(ctx, query, args, v)
	return err
}

func NewEntClient(conf *conf.Bootstrap, m *Metric) (*ent.Client, func(), error) {
	var (
		db  *sql.DB
		err error
	)
	switch conf.Data.Database.Driver {
	case dialect.MySQL:
		db, err = splunksql.Open(conf.Data.Database.Driver, conf.Data.Database.Source)
	case dialect.Postgres:
		db, err = splunksql.Open("pgx", conf.Data.Database.Source)
	}
	if err != nil {
		return nil, nil, fmt.Errorf("failed opening connection to db: %v", err)
	}
	db.SetMaxOpenConns(int(conf.Data.Database.MaxOpen))
	db.SetMaxIdleConns(int(conf.Data.Database.MaxIdle))
	db.SetConnMaxLifetime(conf.Data.Database.ConnMaxLifeTime.AsDuration())
	db.SetConnMaxIdleTime(conf.Data.Database.ConnMaxIdleTime.AsDuration())

	sourceURL, err := url.Parse(conf.Data.Database.Source)
	if err != nil {
		return nil, nil, fmt.Errorf("failed parse source of db: %v", err)
	}
	metricAddr := fmt.Sprintf("%s%s", sourceURL.Host, sourceURL.Path)

	drv := entSql.OpenDB(conf.Data.Database.Driver, db)
	drvWrap := &queryMetricDriver{
		tableNameRe: regexp.MustCompile(` + "`" + `FROM\s+"(\w+)"` + "`" + `),
		durationSec: m.DbDurationSec,
		addr:        metricAddr,
		Driver:      drv,
	}
	ec := ent.NewClient(ent.Driver(drvWrap))

	// Run the auto migration kit.
	if err = ec.Schema.Create(
		context.Background(),
		migrate.WithForeignKeys(false),
	); err != nil {
		return nil, nil, fmt.Errorf("failed creating schema resources: %v", err)
	}

	// Add a global hook that runs on all types and all operations.
	ec.Use(
		MetricsHook(
			WithAddr(metricAddr),
			WithRequestsDuration(m.DbDurationSec),
		),
	)
	ec.Use(
		hook.On(
			IDHook(),
			ent.OpCreate,
		), // 创建时，如果未指定ID，则使用sonyflake自动创建ID
	)
	return ec, func() {
		err = ec.Close()
		if err != nil {
			fmt.Printf("failed closing ent client: %v", err)
		}
	}, nil
}

func NewRedisCmd(conf *conf.Bootstrap, m *Metric) (redis.Cmdable, func(), error) {
	client := redis.NewClient(&redis.Options{
		Addr:         conf.Data.Redis.Addr,
		Password:     conf.Data.Redis.Password,
		DB:           int(conf.Data.Redis.DbIndex),
		DialTimeout:  conf.Data.Redis.DialTimeout.AsDuration(),
		ReadTimeout:  conf.Data.Redis.ReadTimeout.AsDuration(),
		WriteTimeout: conf.Data.Redis.WriteTimeout.AsDuration(),
	})
	// Enable tracing instrumentation.
	err := redisotel.InstrumentTracing(client)
	if err != nil {
		return nil, nil, err
	}
	client.AddHook(mr.NewMetricHook(
		mr.WithAddr(conf.Data.Redis.Addr),
		mr.WithRequestsDuration(m.CacheDurationSec),
	))
	timeout, cancelFunc := context.WithTimeout(context.Background(), time.Second*2)
	defer cancelFunc()
	err = client.Ping(timeout).Err()
	if err != nil {
		return nil, nil, fmt.Errorf("redis connect error: %v", err)
	}
	return client, func() {
		err = client.Close()
		if err != nil {
			fmt.Printf("failed closing redis client: %v", err)
		}
	}, nil
}

func NewCacheFanout(metric *Metric, logger log.Logger) (*fanout.Fanout, func()) {
	cf := fanout.New("cache",
		fanout.WithMetricChanLen(metric.CacheChanLen),
		fanout.WithMetricCount(metric.CacheCount),
		fanout.WithLogger(logger),
	)
	return cf, func() {
		err := cf.Close()
		if err != nil {
			fmt.Printf("failed closing cache fanout: %v", err)
		}
	}
}
`

const dataEntExtTemplate = `
{{- /* delete empty line */ -}}
package data

import (
	"context"
	"fmt"
	"sync"
	"time"

	"{{ (index . 0).ServicePath }}/internal/data/ent"
	"github.com/go-kratos/kratos/v2/metrics"
	"github.com/sony/sonyflake"
)

// WithTx runs callbacks in a transaction.
func WithTx(ctx context.Context, client *ent.Client, fn func(tx *ent.Tx) error) (err error) {
	tx, err := client.Tx(ctx)
	if err != nil {
		return
	}
	defer func() {
		if v := recover(); v != nil {
			err = tx.Rollback()
			if err != nil {
				return
			}
			panic(v)
		}
	}()
	if err = fn(tx); err != nil {
		if rer := tx.Rollback(); rer != nil {
			return rer
		}
		return
	}
	err = tx.Commit()
	return
}

// IDHook Using sonyflake to generate IDs with hook.
func IDHook() ent.Hook {
	var sonyflakeMap sync.Map
	type IDSetter interface {
		SetID(uint64)
	}
	type IDGetter interface {
		ID() (id uint64, exists bool)
	}
	type TypeGetter interface {
		Type() string
	}

	return func(next ent.Mutator) ent.Mutator {
		return ent.MutateFunc(func(ctx context.Context, m ent.Mutation) (ent.Value, error) {
			ig, ok := m.(IDGetter)
			if !ok {
				return nil, fmt.Errorf("mutation %T did not implement IDGetter", m)
			}
			_, exists := ig.ID()
			if !exists {
				is, ok := m.(IDSetter)
				if !ok {
					return nil, fmt.Errorf("mutation %T did not implement IDSetter", m)
				}
				tg, ok := m.(TypeGetter)
				if !ok {
					return nil, fmt.Errorf("mutation %T did not implement TypeGetter", m)
				}
				typ := tg.Type()
				val, ok := sonyflakeMap.Load(typ)
				var idGen *sonyflake.Sonyflake
				if ok {
					idGen = val.(*sonyflake.Sonyflake)
				} else {
					st, _ := time.Parse("2006-01-02", "2022-12-10")
					idGen = sonyflake.NewSonyflake(
						sonyflake.Settings{
							StartTime: st,
						},
					)
					sonyflakeMap.Store(typ, idGen)
				}
				id, err := idGen.NextID()
				if err != nil {
					return nil, err
				}
				is.SetID(id)
			}
			return next.Mutate(ctx, m)
		})
	}
}

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

type options struct {
	// histogram: db_client_requests_duration_sec_bucket{"name", "Addr", "command", "res"}
	requestsDuration metrics.Observer
	Addr             string
}

// MetricsHook Using prometheus to monitor db with hook.
func MetricsHook(opts ...Option) ent.Hook {
	op := options{}
	for _, o := range opts {
		o(&op)
	}
	return func(next ent.Mutator) ent.Mutator {
		return ent.MutateFunc(
			func(ctx context.Context, m ent.Mutation) (val ent.Value, err error) {
				dbOp := ""
				switch m.Op() {
				case ent.OpCreate: // node creation.
					dbOp = "create"
				case ent.OpUpdate: // update nodes by predicate (if any).
					dbOp = "update"
				case ent.OpUpdateOne: // update one node.
					dbOp = "update_one"
				case ent.OpDelete: // delete nodes by predicate (if any).
					dbOp = "delete"
				case ent.OpDeleteOne: // delete one node.
					dbOp = "delete_one"
				}
				if op.requestsDuration != nil {
					startTime := time.Now()
					defer func() {
						res := "ok"
						if err != nil {
							res = fmt.Sprintf("%T", err)
						}
						op.requestsDuration.
							With(m.Type(), op.Addr, dbOp, res).
							Observe(time.Since(startTime).Seconds())
					}()
				}
				val, err = next.Mutate(ctx, m)
				return
			},
		)
	}
}
`

const dataEntSchemaTemplate = `
{{- /* delete empty line */ -}}
package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
    "entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/mixin"
)

// ++++++++++++++++++ Mixin ++++++++++++++++++

// IDMixin to be shared will all different schemas.
type IDMixin struct {
	mixin.Schema
}

// Fields of the Mixin.
func (IDMixin) Fields() []ent.Field {
	return []ent.Field{
		field.Uint64("id").
			Immutable(),
	}
}

// ++++++++++++++++++ entity ++++++++++++++++++
{{- .Ent }}
`

const confTemplate = `
{{- /* delete empty line */ -}}
syntax = "proto3";

package {{ (index . 0).Name }}.conf;

import "google/protobuf/duration.proto";

option go_package = "{{ (index . 0).ServicePath }}/internal/conf;conf";


message ApolloConfig{
  message Apollo {
    string endpoint = 1;
    string app_id = 2;
    string cluster = 3;
    string namespace = 4;
    string secret = 5;
  }
  Apollo apollo = 1;
}

message AppInfo {
  string name = 1;
  string version = 2;
  string flag_conf = 3;
  string id = 4;
}

message Bootstrap {
  Trace trace = 1;
  Server server = 2;
  Data data = 3;
  Registry registry = 4;
  Auth auth = 5;
  Log log = 6;
}

message Trace {
  string endpoint = 1;
}

message Server {
  message HTTP {
    string network = 1;
    string addr = 2;
    google.protobuf.Duration timeout = 3;
  }
  message GRPC {
    string network = 1;
    string addr = 2;
    google.protobuf.Duration timeout = 3;
  }
  HTTP http = 1;
  GRPC grpc = 2;
}

message Data {
  message Database {
    string driver = 1;
    string source = 2;
    int64 max_open = 3;
    int64 max_idle = 4;
    google.protobuf.Duration conn_max_life_time = 5;
    google.protobuf.Duration conn_max_idle_time = 6;
  }
  message Redis {
    string addr = 1;
    string password = 2;
    int64  db_index = 3;
    google.protobuf.Duration  dial_timeout = 4;
    google.protobuf.Duration  read_timeout = 5;
    google.protobuf.Duration  write_timeout = 6;
  }
  Database database = 1;
  Redis redis = 2;
}

message Registry {
  message Consul {
    string address = 1;
    string scheme = 2;
  }
  Consul consul = 1;
}

message Auth {
  string key = 1;
}

message Log {
  enum Level {
    DEBUG = 0;
    INFO = 1;
    WARN = 2;
    ERROR = 3;
  }

  enum Encoding {
    JSON = 0;
    CONSOLE = 1;
  }

  message Sampling {
    // This will log the first initial log entries with the same level and message
    // in a one second interval as-is. Following that, it will allow through
    // every 5th log entry with the same level and message in that interval.
    int32 initial = 1;
    // If thereafter is zero, the Core will drop all log entries after the first N
    // in that interval.
    int32 thereafter = 2;
  }

  message Output {
    message Rotate {
      // max_size is the maximum size in megabytes of the log file before it gets
      // rotated. It defaults to 100 megabytes.
      int32 max_size = 2;
      // max_age is the maximum number of days to retain old log files based on the
      // timestamp encoded in their filename.  Note that a day is defined as 24
      // hours and may not exactly correspond to calendar days due to daylight
      // savings, leap seconds, etc. The default is not to remove old log files
      // based on age.
      int32 max_age = 3;
      // max_backups is the maximum number of old log files to retain.  The default
      // is to retain all old log files (though max_age may still cause them to get
      // deleted.)
      int32 max_backups = 4;
      // compress determines if the rotated log files should be compressed
      // using gzip. The default is not to perform compression.
      bool compress = 5;
    }

    // Since it's common to write logs to the local filesystem, URLs without a
    // scheme (e.g., "/var/log/foo.log") are treated as local file paths. Without
    // a scheme, the special paths "stdout" and "stderr" are interpreted as
    // os.Stdout and os.Stderr. When specified without a scheme, relative file
    // paths also work.
    string path = 1;

    // The following option are only available if the path is to the local
    // file system
    Rotate rotate = 2;
  }

  Level level = 1;
  Encoding encoding = 2;
  // default sampling.is_initial = 100, sampling.thereafter = 100
  Sampling sampling = 3;
  // default output_paths.path = "stderr"
  repeated Output output_paths = 4;
}
`

const serverServerTemplate = `
{{- /* delete empty line */ -}}
package server

import (
	"github.com/google/wire"
)

// ProviderSet is server providers.
var ProviderSet = wire.NewSet(NewGRPCServer, NewHTTPServer)
`

const serverGRPCTemplate = `
{{- /* delete empty line */ -}}
package server

import (
	{{ (index . 0).Version }} "{{ (index . 0).PbPath }}"
	"{{ (index . 0).ServicePath }}/internal/conf"
	"{{ (index . 0).ServicePath }}/internal/data"
	"{{ (index . 0).ServicePath }}/internal/service"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware/logging"
	"github.com/go-kratos/kratos/v2/middleware/metrics"
	"github.com/go-kratos/kratos/v2/middleware/ratelimit"
	"github.com/go-kratos/kratos/v2/middleware/recovery"
	"github.com/go-kratos/kratos/v2/middleware/tracing"
	"github.com/go-kratos/kratos/v2/middleware/validate"
	"github.com/go-kratos/kratos/v2/transport/grpc"
	"go.opentelemetry.io/otel/trace"
)

// NewGRPCServer new a gRPC server.
func NewGRPCServer(
	bc *conf.Bootstrap,
	logger log.Logger,
	m *data.Metric,
{{- range $index, $element := .}}
{{- if eq $index 0 }}
	s *service.{{ $element.Service }},
{{- else }}
	s{{ $index }} *service.{{ $element.Service }},
{{- end }}
{{- end }}
	_ trace.TracerProvider, // otel.SetTracerProvider(provider) instead, but need to declare to wire injection
) *grpc.Server {
	cs := bc.Server
	// ca := bc.Auth
	opts := []grpc.ServerOption{
		grpc.Middleware(
			recovery.Recovery(),
			tracing.Server(),
			metrics.Server(
				metrics.WithSeconds(m.DurationSec),
				metrics.WithRequests(m.CodeTotal),
			),
			recovery.Recovery(),
			logging.Server(logger),
			ratelimit.Server(),
			validate.Validator(),
			//jwt.Server(func(token *jwtV4.Token) (interface{}, error) {
			//	return []byte(ca.Key), nil
			//}),
		),
	}
	if cs.Grpc.Network != "" {
		opts = append(opts, grpc.Network(cs.Grpc.Network))
	}
	if cs.Grpc.Addr != "" {
		opts = append(opts, grpc.Address(cs.Grpc.Addr))
	}
	if cs.Grpc.Timeout != nil {
		opts = append(opts, grpc.Timeout(cs.Grpc.Timeout.AsDuration()))
	}
	srv := grpc.NewServer(opts...)
{{- range $index, $element := .}}
{{- if eq $index 0 }}
	{{ $element.Version }}.Register{{ $element.Service }}Server(srv, s)
{{- else }}
	{{ $element.Version }}.Register{{ $element.Service }}Server(srv, s{{ $index }})
{{- end }}
{{- end }}
	return srv
}
`

const serverHTTPTemplate = `
{{- /* delete empty line */ -}}
package server

import (
{{- if eq (index . 0).ServiceType "Interface" }}
	{{ (index . 0).Version }} "{{ (index . 0).PbPath }}"
{{- else if eq (index . 0).ServiceType "Admin" }}
	{{ (index . 0).Version }} "{{ (index . 0).PbPath }}"
{{- end }}
	"{{ (index . 0).ServicePath }}/internal/conf"
	"{{ (index . 0).ServicePath }}/internal/data"
{{- if eq (index . 0).ServiceType "Interface" }}
	"{{ (index . 0).ServicePath }}/internal/service"
{{- else if eq (index . 0).ServiceType "Admin" }}
	"{{ (index . 0).ServicePath }}/internal/service"
{{- end }}
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware/logging"
	"github.com/go-kratos/kratos/v2/middleware/metrics"
	"github.com/go-kratos/kratos/v2/middleware/ratelimit"
	"github.com/go-kratos/kratos/v2/middleware/recovery"
	"github.com/go-kratos/kratos/v2/middleware/tracing"
	"github.com/go-kratos/kratos/v2/middleware/validate"
	"github.com/go-kratos/kratos/v2/transport/http"
	"github.com/gorilla/handlers"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/otel/trace"
)

// NewHTTPServer new an HTTP server.
func NewHTTPServer(
	bc *conf.Bootstrap,
	logger log.Logger,
	m *data.Metric,
{{- range $index, $element := .}}
{{- if eq $element.ServiceType "Interface" }}
{{- if eq $index 0 }}
	s *service.{{ $element.Service }},
{{- else }}
	s{{ $index }} *service.{{ $element.Service }},
{{- end }}
{{- else if eq $element.ServiceType "Admin" }}
{{- if eq $index 0 }}
	s *service.{{ $element.Service }},
{{- else }}
	s{{ $index }} *service.{{ $element.Service }},
{{- end }}
{{- end }}
{{- end }}
	_ trace.TracerProvider, // otel.SetTracerProvider(provider) instead, but need to declare to wire injection
) *http.Server {
	cs := bc.Server
	// ca := bc.Auth
	opts := []http.ServerOption{
		http.Middleware(
			recovery.Recovery(),
			tracing.Server(),
			metrics.Server(
				metrics.WithSeconds(m.DurationSec),
				metrics.WithRequests(m.CodeTotal),
			),
			recovery.Recovery(),
			logging.Server(logger),
			ratelimit.Server(),
			validate.Validator(),
			//jwt.Server(func(token *jwtV4.Token) (interface{}, error) {
			//	return []byte(ca.Key), nil
			//}),
		),
		http.Filter(handlers.CORS(
			handlers.AllowedHeaders([]string{"X-Requested-With", "Content-Type", "Authorization"}),
			handlers.AllowedMethods([]string{"GET", "POST", "PUT", "HEAD", "OPTIONS"}),
			handlers.AllowedOrigins([]string{"*"}),
		)),
	}
	if cs.Http.Network != "" {
		opts = append(opts, http.Network(cs.Http.Network))
	}
	if cs.Http.Addr != "" {
		opts = append(opts, http.Address(cs.Http.Addr))
	}
	if cs.Http.Timeout != nil {
		opts = append(opts, http.Timeout(cs.Http.Timeout.AsDuration()))
	}
	srv := http.NewServer(opts...)
	srv.Handle("/metrics", promhttp.Handler())
{{- range $index, $element := .}}
{{- if eq $element.ServiceType "Interface" }}
{{- if eq $index 0 }}
	{{ $element.Version }}.Register{{ $element.Service }}HTTPServer(srv, s)
{{- else }}
	{{ $element.Version }}.Register{{ $element.Service }}HTTPServer(srv, s{{ $index }})
{{- end }}
{{- else if eq $element.ServiceType "Admin" }}
{{- if eq $index 0 }}
	{{ $element.Version }}.Register{{ $element.Service }}HTTPServer(srv, s)
{{- else }}
	{{ $element.Version }}.Register{{ $element.Service }}HTTPServer(srv, s{{ $index }})
{{- end }}
{{- end }}
{{- end }}
	return srv
}
`

const testDockerTemplate = `
{{- /* delete empty line */ -}}
version: "3.7"

services:
  db:
    image: postgres
    environment:
      POSTGRES_PASSWORD: example
    ports:
      - "12211:5432"
#  db:
#    image: mysql
#    environment:
#      MYSQL_ROOT_PASSWORD: example
#      MYSQL_DATABASE: test
#    ports:
#      - "12213:3306"

  redis:
    image: redis
    ports:
      - "12212:6379"
`

const testServiceTemplate = `
{{- /* delete empty line */ -}}
package test

import (
	"os"
	"testing"

	"{{ (index . 0).ServicePath }}/internal/conf"
	"{{ (index . 0).ServicePath }}/internal/service"
	"github.com/tianping526/kit/testing/lich"
)

{{ range $index, $element := . -}}
{{ if eq $index 0 -}}
var sv *service.{{ $element.Service }}
{{ else -}}
var sv{{ $index }} *service.{{ $element.Service }}
{{- end }}
{{- end }}
func TestMain(m *testing.M) {
	var err error
	code := 2212

	defer func() {
		os.Exit(code)
	}()

	//err = flag.Set("f", "../.testdata/docker-compose.yaml")
	//if err != nil {
	//	panic(err)
	//}
	//flag.Parse()

	err = lich.Teardown()
	if err != nil {
		panic(err)
	}

	if err = lich.Setup(); err != nil {
		panic(err)
	}
	defer func() {
		err = lich.Teardown()
		if err != nil {
			panic(err)
		}
	}()

	var appInfo conf.AppInfo
	appInfo.Id = "test"
	appInfo.Name = "{{ (index . 0).Name }}"
	appInfo.Version = ""
	appInfo.FlagConf = "../configs"
{{ range $index, $element := . }}
{{- if eq $index 0 }}
	var cleanup func()
	sv, cleanup, err = wireService(&appInfo)
	if err != nil {
		panic(err)
	}
	defer cleanup()
{{- else }}
	var cleanup{{ $index }} func()
	sv{{ $index }}, cleanup{{ $index }}, err = wireService{{ $index }}(&appInfo)
	if err != nil {
		panic(err)
	}
	defer cleanup{{ $index }}()
{{- end }}
{{- end }}

	code = m.Run()
}
`

const testWireTemplate = `
//go:build wireinject
// +build wireinject

// The build tag makes sure the stub is not built in the final build.

package test

import (
	"github.com/google/wire"
	"{{ (index . 0).ServicePath }}/internal/biz"
	"{{ (index . 0).ServicePath }}/internal/conf"
	"{{ (index . 0).ServicePath }}/internal/data"
	"{{ (index . 0).ServicePath }}/internal/service"
)

{{ range $index, $element := . -}}
func wireService{{ if ne $index 0 }}{{ $index }}{{ end }}(*conf.AppInfo) (*service.{{ $element.Service -}}
, func(), error) {
	panic(wire.Build(data.ProviderSet, biz.ProviderSet, service.ProviderSet))
}
{{ end -}}
`
