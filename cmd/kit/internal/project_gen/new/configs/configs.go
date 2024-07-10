package configs

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
)

type Configs struct {
	Path string
}

// Generate generate a configs template.
func (c *Configs) Generate() error {
	err := c.generate("apollo.yaml", apolloTemplate, false)
	if err != nil {
		return err
	}
	return c.generate("service.yaml", serviceTemplate, false)
}

func (c *Configs) generate(name, tpl string, code bool) error {
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
	to := filepath.Join(wd, c.Path, "configs")
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

func (c *Configs) execute(tpl string) ([]byte, error) {
	buf := new(bytes.Buffer)
	tmpl, err := template.New("configs").Parse(tpl)
	if err != nil {
		return nil, err
	}
	if err := tmpl.Execute(buf, c); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// CmdConfigs represents the Configs command.
var CmdConfigs = &cobra.Command{
	Use:   "configs",
	Short: "new a configs template",
	Long:  "new a configs template. Example: kit new configs xip.event.interface.v1 [code dir]",
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
	pkgPath := filepath.Join(codeDir, strings.Join(parts[:len(parts)-1], "/"))
	c := Configs{
		Path: pkgPath,
	}
	if err := c.Generate(); err != nil {
		fmt.Println(err)
		return
	}
}
