package cmd

const mainTemplate = `
{{- /* delete empty line */ -}}
package main

import (
	"flag"
	_ "net/http/pprof"
	"os"

	"{{ .ModName }}/{{ .Path }}/internal/conf"
	"github.com/go-kratos/kratos/v2"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/registry"
	"github.com/go-kratos/kratos/v2/transport/grpc"
	"github.com/go-kratos/kratos/v2/transport/http"
	_ "go.uber.org/automaxprocs"
)

// go build -ldflags "-X main.Version=x.y.z"
var (
	// Name is the name of the compiled software.
	Name = "{{ .Name }}"
	// Version is the version of the compiled software.
	Version string
	// flagConf is the config flag.
	flagConf string

	id, _ = os.Hostname()
)

func init() {
	flag.StringVar(&flagConf, "conf", "../../configs", "config path, eg: -conf config.yaml")
}

func newApp(logger log.Logger, hs *http.Server, gs *grpc.Server, rr registry.Registrar) *kratos.App {
	return kratos.New(
		kratos.ID(id),
		kratos.Name(Name),
		kratos.Version(Version),
		kratos.Metadata(map[string]string{}),
		kratos.Logger(logger),
		kratos.Server(
			hs,
			gs,
		),
		kratos.Registrar(rr),
	)
}

func main() {
	flag.Parse()

	var appInfo conf.AppInfo
	appInfo.Id = id
	appInfo.Name = Name
	appInfo.Version = Version
	appInfo.FlagConf = flagConf
	app, cleanup, err := wireApp(&appInfo)
	if err != nil {
		panic(err)
	}
	defer cleanup()

	// start and wait for stop signal
	if err := app.Run(); err != nil {
		panic(err)
	}
}
`

const wireTemplate = `
{{- /* delete empty line */ -}}
//go:build wireinject
// +build wireinject

// The build tag makes sure the stub is not built in the final build.

package main

import (
	"github.com/go-kratos/kratos/v2"
	"github.com/google/wire"
	"{{ .ModName }}/{{ .Path }}/internal/biz"
	"{{ .ModName }}/{{ .Path }}/internal/conf"
	"{{ .ModName }}/{{ .Path }}/internal/data"
	"{{ .ModName }}/{{ .Path }}/internal/server"
	"{{ .ModName }}/{{ .Path }}/internal/service"
)

// wireApp init kratos application.
func wireApp(*conf.AppInfo) (*kratos.App, func(), error) {
	panic(wire.Build(server.ProviderSet, data.ProviderSet, biz.ProviderSet, service.ProviderSet, newApp))
}
`
