package add

import (
	"bytes"
	"fmt"
	"go/format"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"text/template"

	"github.com/spf13/cobra"

	"github.com/tianping526/kit/cmd/kit/internal/project_gen/proto/common"
)

// CmdAdd represents the add command.
var CmdAdd = &cobra.Command{
	Use:   "add",
	Short: "Add a proto API template",
	Long:  "Add a proto API template. Example: kit proto add xip.event.interface.v1 [proto dir]",
	Run:   run,
}

// Proto is a proto generator.
type Proto struct {
	Name        string
	Path        string
	HTTPPath    string
	Service     string
	ServiceType string
	Package     string
	GoPackage   string
	JavaPackage string
}

func run(_ *cobra.Command, args []string) {
	// kit add xip.event.interface.v1
	if len(args) == 0 {
		fmt.Println("need to specify service name.  Example: xip.event.interface.v1")
	}
	input := args[0]
	re := regexp.MustCompile(`^((\w|_)+\.){3}v\d+$`)
	mr := re.Match([]byte(input))
	if !mr {
		fmt.Println("service name err. Example: xip.event.interface.v1")
	}
	parts := strings.Split(input, ".")
	httpPath := filepath.Join("/", parts[len(parts)-1], strings.Join(parts[:len(parts)-2], "/"))
	dir := "xapis/api/"
	if len(args) >= 2 { //nolint:mnd
		specDir := strings.TrimSpace(args[1])
		if specDir != "" {
			dir = specDir
		}
	}
	protoDir := filepath.Join(dir, strings.Join(parts, "/"))
	pkgName := input
	name := strings.Join(parts, "_")

	p := &Proto{
		Name:        common.SnakeString(name),
		HTTPPath:    httpPath,
		Path:        protoDir,
		Package:     pkgName,
		GoPackage:   goPackage(protoDir),
		JavaPackage: javaPackage(pkgName),
		Service:     common.CamelString(parts[1]),
		ServiceType: strings.ToUpper(parts[2][:1]) + parts[2][1:],
	}
	if err := p.Generate(); err != nil {
		fmt.Println(err)
		return
	}
}

func goPackage(protoPath string) string {
	s := strings.Split(protoPath, "/")
	return strings.Join([]string{filepath.Join(common.ModName(), protoPath), s[len(s)-1]}, ";")
}

func javaPackage(name string) string {
	return name
}

// Generate generate a proto template.
func (p *Proto) Generate() error {
	err := p.generate(p.Name+".proto", protoTemplate, false)
	if err != nil {
		return err
	}
	return p.generate(p.Name+"_error.proto", errorProtoTemplate, false)
}

func (p *Proto) execute(tpl string) ([]byte, error) {
	buf := new(bytes.Buffer)
	tmpl, err := template.New("proto").Parse(tpl)
	if err != nil {
		return nil, err
	}
	if err := tmpl.Execute(buf, p); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (p *Proto) generate(name, tpl string, code bool) error {
	body, err := p.execute(tpl)
	if err != nil {
		return err
	}
	if code {
		body, err = format.Source(body)
		if err != nil {
			return err
		}
	}
	wd, err := os.Getwd()
	if err != nil {
		return err
	}
	to := filepath.Join(wd, p.Path)
	if _, err = os.Stat(to); os.IsNotExist(err) {
		if err = os.MkdirAll(to, 0o700); err != nil {
			return err
		}
	}
	name = filepath.Join(to, name)
	if _, err = os.Stat(name); !os.IsNotExist(err) {
		return fmt.Errorf("%s already exists", p.Name)
	}
	return os.WriteFile(name, body, 0o644)
}
