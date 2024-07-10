package server

import (
	"bytes"
	"fmt"
	"go/format"
	"log"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/tianping526/kit/cmd/kit/internal/project_gen/proto/common"
)

type MethodType uint8

const (
	unaryType          MethodType = 1
	twoWayStreamsType  MethodType = 2
	requestStreamsType MethodType = 3
	returnsStreamsType MethodType = 4
)

// Service is a proto service.
type Service struct {
	Package             string
	Service             string
	ServiceName         string
	ServiceType         string
	ServiceInternalName string
	ServicePath         string
	PbPath              string
	Name                string
	Methods             []*Method
	GoogleEmpty         bool
	Version             string

	DO      string
	Imports []string
	Ent     string

	UseIO      bool
	UseContext bool
}

// Method is a proto method.
type Method struct {
	Service string
	Name    string
	Request string
	Reply   string

	// type: unary or stream
	Type MethodType

	ReqCopy   string
	RepCopy   string
	RepDoName string
}

func (s *Service) execute() ([]byte, error) {
	buf := new(bytes.Buffer)
	tmpl, err := template.New("service").Funcs(map[string]any{
		"dtoCovertDoName": dtoCovertDoName,
		"toInternalName":  toInternalName,
	}).Parse(serviceTemplate)
	if err != nil {
		return nil, err
	}
	if err := tmpl.Execute(buf, s); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func getFileNameFormService(service string) string {
	target := getServiceName(service)
	return common.SnakeString(target) + ".go"
}

func getServiceName(service string) string {
	target := service
	suffix := []string{"Interface", "Admin", "Service", "Job"}
	for _, s := range suffix {
		if strings.HasSuffix(service, s) {
			target = strings.TrimSuffix(service, s)
			break
		}
	}
	return target
}

func getServiceInternalName(serviceName string) string {
	return strings.ToLower(string(serviceName[0])) + serviceName[1:]
}

func genService(codePath string, res []*Service) {
	targetDir := filepath.Join(codePath, "internal/service")
	if _, err := os.Stat(targetDir); os.IsNotExist(err) {
		if err := os.MkdirAll(targetDir, 0o700); err != nil {
			log.Fatal(err)
		}
	}
	for _, s := range res {
		to := filepath.Join(targetDir, getFileNameFormService(s.Service))
		if _, err := os.Stat(to); !os.IsNotExist(err) {
			_, _ = fmt.Fprintf(os.Stderr, "%s already exists: %s\n", s.Service, to)
			continue
		}
		b, err := s.execute()
		if err != nil {
			log.Fatal(err)
		}
		b, err = format.Source(b)
		if err != nil {
			log.Fatal(err)
		}
		if err := os.WriteFile(to, b, 0o644); err != nil {
			log.Fatal(err)
		}
		fmt.Println(to)
	}
	to := filepath.Join(targetDir, "README.md")
	genFile(to, serviceReadmeTemplate, res, map[string]any{}, false)
	to = filepath.Join(targetDir, "service.go")
	genFile(to, serviceServiceTemplate, res, map[string]any{}, true)
}

func genBiz(codePath string, res []*Service) {
	targetDir := filepath.Join(codePath, "internal/biz")
	if _, err := os.Stat(targetDir); os.IsNotExist(err) {
		if err := os.MkdirAll(targetDir, 0o700); err != nil {
			log.Fatal(err)
		}
	}
	for _, s := range res {
		to := filepath.Join(targetDir, getFileNameFormService(s.Service))
		if _, err := os.Stat(to); !os.IsNotExist(err) {
			_, _ = fmt.Fprintf(os.Stderr, "%s already exists: %s\n", s.Service, to)
			continue
		}
		genFile(to, bizTemplate, s, map[string]any{
			"dtoCovertDoName": dtoCovertDoName,
		}, true)
	}
	to := filepath.Join(targetDir, "README.md")
	genFile(to, bizReadmeTemplate, res, map[string]any{}, false)
	to = filepath.Join(targetDir, "biz.go")
	genFile(to, bizBizTemplate, res, map[string]any{}, true)
}

func genData(codePath string, res []*Service) {
	entSchemaDir := filepath.Join(codePath, "internal/data/ent/schema")
	if _, err := os.Stat(entSchemaDir); os.IsNotExist(err) {
		if err := os.MkdirAll(entSchemaDir, 0o700); err != nil {
			log.Fatal(err)
		}
	}
	targetDir := filepath.Join(codePath, "internal/data")
	for _, s := range res {
		to := filepath.Join(targetDir, getFileNameFormService(s.Service))
		_, err := os.Stat(to)
		if os.IsNotExist(err) {
			genFile(to, dataTemplate, s, map[string]any{
				"dtoCovertDoName": dtoCovertDoName,
			}, true)
		} else {
			_, _ = fmt.Fprintf(os.Stderr, "already exists: %s\n", to)
		}
		to = filepath.Join(entSchemaDir, getFileNameFormService(s.Service))
		_, err = os.Stat(to)
		if os.IsNotExist(err) {
			genFile(to, dataEntSchemaTemplate, s, map[string]any{}, true)
		} else {
			_, _ = fmt.Fprintf(os.Stderr, "already exists: %s\n", to)
		}
	}
	to := filepath.Join(targetDir, "README.md")
	genFile(to, dataReadmeTemplate, res, map[string]any{}, false)
	to = filepath.Join(targetDir, "data.go")
	genFile(to, dataDataTemplate, res, map[string]any{}, true)
	to = filepath.Join(targetDir, "ent_ext.go")
	genFile(to, dataEntExtTemplate, res, map[string]any{}, true)
}

func genConf(codePath string, res []*Service) {
	targetDir := filepath.Join(codePath, "internal/conf")
	if _, err := os.Stat(targetDir); os.IsNotExist(err) {
		if err := os.MkdirAll(targetDir, 0o700); err != nil {
			log.Fatal(err)
		}
	}
	to := filepath.Join(targetDir, "conf.proto")
	genFile(to, confTemplate, res, map[string]any{}, false)
}

func genServer(codePath string, res []*Service) {
	targetDir := filepath.Join(codePath, "internal/server")
	if _, err := os.Stat(targetDir); os.IsNotExist(err) {
		if err := os.MkdirAll(targetDir, 0o700); err != nil {
			log.Fatal(err)
		}
	}
	to := filepath.Join(targetDir, "server.go")
	genFile(to, serverServerTemplate, res, map[string]any{}, true)
	to = filepath.Join(targetDir, "grpc.go")
	genFile(to, serverGRPCTemplate, res, map[string]any{}, true)
	to = filepath.Join(targetDir, "http.go")
	genFile(to, serverHTTPTemplate, res, map[string]any{}, true)
}

func genTest(codePath string, res []*Service) {
	targetDir := filepath.Join(codePath, "test")
	if _, err := os.Stat(targetDir); os.IsNotExist(err) {
		if err := os.MkdirAll(targetDir, 0o700); err != nil {
			log.Fatal(err)
		}
	}
	to := filepath.Join(targetDir, "docker-compose.yaml")
	genFile(to, testDockerTemplate, res, map[string]any{}, false)
	to = filepath.Join(targetDir, "service_test.go")
	genFile(to, testServiceTemplate, res, map[string]any{}, true)
	to = filepath.Join(targetDir, "wire.go")
	genFile(to, testWireTemplate, res, map[string]any{}, true)
}

func genFile(to, tpl string, data any, fs template.FuncMap, code bool) {
	_, err := os.Stat(to)
	if os.IsNotExist(err) {
		tmpl, err := template.New("file").Funcs(fs).Parse(tpl)
		if err != nil {
			log.Fatal(err)
		}
		buf := new(bytes.Buffer)
		if err = tmpl.Execute(buf, data); err != nil {
			log.Fatal(err)
		}
		body := buf.Bytes()
		if code {
			body, err = format.Source(body)
			if err != nil {
				log.Fatal(err)
			}
		}
		if err := os.WriteFile(to, body, 0o644); err != nil {
			log.Fatal(err)
		}
		fmt.Println(to)
	} else {
		_, _ = fmt.Fprintf(os.Stderr, "already exists: %s\n", to)
	}
}
