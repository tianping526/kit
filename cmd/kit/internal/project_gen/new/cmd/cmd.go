package cmd

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

type Cmd struct {
	Name    string
	Path    string
	ModName string
}

// Generate generate a cmd template.
func (c *Cmd) Generate() error {
	err := c.generate("main.go", mainTemplate, true)
	if err != nil {
		return err
	}
	return c.generate("wire.go", wireTemplate, true)
}

func (c *Cmd) generate(name, tpl string, code bool) error {
	body, err := c.execute(tpl)
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
	to := filepath.Join(wd, c.Path, "cmd/server")
	if _, err = os.Stat(to); os.IsNotExist(err) {
		if err = os.MkdirAll(to, 0o700); err != nil {
			return err
		}
	}
	name = filepath.Join(to, name)
	if _, err = os.Stat(name); !os.IsNotExist(err) {
		fmt.Printf("%s already exists\n", name)
		return nil
	}
	err = os.WriteFile(name, body, 0o644)
	if err != nil {
		return err
	}
	fmt.Println(name)
	return nil
}

func (c *Cmd) execute(tpl string) ([]byte, error) {
	buf := new(bytes.Buffer)
	tmpl, err := template.New("cmd").Parse(tpl)
	if err != nil {
		return nil, err
	}
	if err := tmpl.Execute(buf, c); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// CmdCmd represents the cmd command.
var CmdCmd = &cobra.Command{
	Use:   "cmd",
	Short: "new a cmd template",
	Long:  "new a cmd template. Example: kit new cmd xip.event.interface.v1 [code dir]",
	Run:   run,
}

func run(_ *cobra.Command, args []string) {
	if len(args) == 0 {
		fmt.Println("need to specify service name.  Example: xip.event.interface.v1 [code dir]")
	}
	input := args[0]
	re := regexp.MustCompile(`^((\w|_)+\.){3}v\d+$`)
	mr := re.Match([]byte(input))
	if !mr {
		fmt.Println("service name err. Example: xip.event.interface.v1")
	}
	parts := strings.Split(input, ".")

	codeDir := "app"
	if len(args) > 1 { //nolint:mnd
		codeDir = args[1]
	}

	name := strings.Join(parts[:len(parts)-1], ".")
	pkgPath := filepath.Join(codeDir, strings.Join(parts[:len(parts)-1], "/"))
	c := Cmd{
		Name:    name,
		Path:    pkgPath,
		ModName: common.ModName(),
	}
	if err := c.Generate(); err != nil {
		fmt.Println(err)
		return
	}
}
