package trivial

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

type Trivial struct {
	Name string
	Path string
}

// Generate generate a trivial template.
func (c *Trivial) Generate() error {
	err := c.generate("generate.go", generateTemplate, true)
	if err != nil {
		return err
	}
	return c.generate("README.md", readmeTemplate, false)
}

func (c *Trivial) generate(name, tpl string, code bool) error {
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
	to := filepath.Join(wd, c.Path)
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

func (c *Trivial) execute(tpl string) ([]byte, error) {
	buf := new(bytes.Buffer)
	tmpl, err := template.New("trivial").Parse(tpl)
	if err != nil {
		return nil, err
	}
	if err := tmpl.Execute(buf, c); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// CmdTrivial represents the trivial command.
var CmdTrivial = &cobra.Command{
	Use:   "trivial",
	Short: "new a trivial template",
	Long:  "new a trivial template. Example: kit new trivial xip.event.interface.v1 [code dir]",
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
	c := Trivial{
		Name: name,
		Path: pkgPath,
	}
	if err := c.Generate(); err != nil {
		fmt.Println(err)
		return
	}
}
