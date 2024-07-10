package trivial

const generateTemplate = `
{{- /* delete empty line */ -}}
package generate

//go:generate buf generate --template {"version":"v1","plugins":[{"plugin":"go","out":".","opt":["paths=source_relative"]}]}
//go:generate go run -mod=mod entgo.io/ent/cmd/ent generate ./internal/data/ent/schema --feature sql/lock
//go:generate go run -mod=mod github.com/google/wire/cmd/wire ./cmd/server
//go:generate go run -mod=mod github.com/google/wire/cmd/wire ./test
`

const readmeTemplate = `
{{- /* delete empty line */ -}}
# {{ .Name }}
`
