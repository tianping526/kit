package rc

import (
	"bytes"
	"flag"
	"go/ast"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"text/template"

	"github.com/spf13/cobra"

	"github.com/tianping526/kit/cmd/kit/internal/pkg"
)

var (
	encode         = flag.String("encode", "", "encode type: json/pb/raw/gob/gzip")
	rcType         = flag.String("type", "", "type: get/set/del/only_add")
	key            = flag.String("key", "", "key name method")
	expire         = flag.String("expire", "", "expire time code")
	structName     = flag.String("rc_struct_name", "repo", "struct name")
	batchSize      = flag.Int("rc_batch", 0, "redis cache batch size")
	batchErr       = flag.String("rc_batch_err", "break", "batch err to continue or break")
	maxGroup       = flag.Int("rc_max_group", 0, "max group size")
	checkNullCode  = flag.String("rc_check_null_code", "", "check null code")
	nullExpire     = flag.String("null_expire", "", "null cache expire time code")
	casCode        = flag.String("cas_code", "", "redis lua cas code")
	casVersionCode = flag.String("cas_version_code", "", "redis lua cas version code")

	rcValidTypes   = []string{"set", "del", "get", "only_add"}
	rcValidPrefix  = []string{"set", "del", "get", "cache", "add"}
	optionNamesMap = map[string]bool{
		"rc_batch": true, "rc_max_group": true, "encode": true, "type": true,
		"key": true, "expire": true, "rc_batch_err": true, "rc_struct_name": true,
		"rc_check_null_code": true, "null_expire": true, "cas_code": true, "cas_version_code": true,
	}
	simpleTypes = []string{
		"int", "int8", "int16", "int32", "int64", "float32", "float64",
		"uint", "uint8", "uint16", "uint32", "uint64", "bool", stringType, "[]byte",
	}
	lenTypes = []string{"[]", "map"}
)

const (
	_interfaceName = "_rc"
	_multiTpl      = 1
	_singleTpl     = 2
	_noneTpl       = 3
	_typeGet       = "get"
	_typeSet       = "set"
	_typeDel       = "del"
	_typeAdd       = "only_add"
	_ex            = 2
	_pn            = 2
	stringType     = "string"
)

func resetFlag() {
	*encode = ""
	*rcType = ""
	*batchSize = 0
	*maxGroup = 0
	*batchErr = "break"
	*checkNullCode = ""
	*nullExpire = ""
	*casCode = ""
	*casVersionCode = ""
	*structName = "repo"
}

// options options
type options struct {
	name        string
	keyType     string
	ValueType   string
	template    int
	SimpleValue bool
	// int float 类型
	GetSimpleValue bool
	// string, []byte类型
	GetDirectValue     bool
	ConvertValue2Bytes string
	ConvertBytes2Value string
	GoValue            bool
	ImportPackage      string
	importPackages     []string
	Args               string
	PkgName            string
	ExtraArgsType      string
	ExtraArgs          string
	RCType             string
	KeyMethod          string
	ExpireCode         string
	Encode             string
	UseRedis           bool
	OriginValueType    string
	UseStrConv         bool
	Comment            string
	GroupSize          int
	MaxGroup           int
	EnableBatch        bool
	BatchErrBreak      bool
	LenType            bool
	PointType          bool
	StructName         string
	CheckNullCode      string
	ExpireNullCode     string
	EnableNullCode     bool
	CasCode            string
	CasVersionCode     string
}

func getOptions(opt *options, comments []string) {
	os.Args = []string{os.Args[0]}
	for _, comment := range comments {
		if regexp.MustCompile(`\s+//\s*(\w+\s)*rc:.+`).Match([]byte(comment)) {
			args := strings.Split(pkg.RegexpReplace(`//\s*(\w+\s)*rc:(?P<arg>.+)`, comment, "$arg"), " ")
			for _, arg := range args {
				arg = strings.TrimSpace(arg)
				if arg != "" {
					// validate option name
					argName := pkg.RegexpReplace(`-(?P<name>[\w_-]+)=.+`, arg, "$name")
					if !optionNamesMap[argName] {
						log.Fatalf("选项:%s 不存在 请检查拼写\n", argName)
					}
					os.Args = append(os.Args, arg)
				}
			}
		}
	}
	resetFlag()
	flag.Parse()
	if *rcType != "" {
		opt.RCType = *rcType
	}
	if *key != "" {
		opt.KeyMethod = *key
	}
	if *expire != "" {
		opt.ExpireCode = *expire
	}
	opt.EnableBatch = (*batchSize != 0) && (*maxGroup != 0)
	opt.BatchErrBreak = *batchErr == "break"
	opt.GroupSize = *batchSize
	opt.MaxGroup = *maxGroup
	opt.StructName = *structName
	opt.CheckNullCode = *checkNullCode
	if *nullExpire != "" {
		opt.ExpireNullCode = *nullExpire
	}
	if *casCode != "" {
		opt.CasCode = *casCode
	}
	if *casVersionCode != "" {
		opt.CasVersionCode = *casVersionCode
	}
	if opt.CheckNullCode != "" {
		opt.EnableNullCode = true
	}
}

func getTypeFromPrefix(opt *options, params []*ast.Field, s *pkg.Source) {
	if opt.RCType == "" {
		for _, t := range rcValidPrefix {
			if strings.HasPrefix(strings.ToLower(opt.name), t) {
				if t == "add" {
					t = _typeSet
				}
				opt.RCType = t
				break
			}
		}
		if opt.RCType == "" {
			log.Fatalln(opt.name + "请指定方法类型(type=get/set/del...)")
		}
	}
	if opt.RCType == "cache" {
		opt.RCType = _typeGet
	}
	if len(params) == 0 {
		log.Fatalln(opt.name + "参数不足")
	}
	for _, p := range params {
		if len(p.Names) > 1 {
			log.Fatalln(opt.name + "不支持省略类型 请写全声明中的字段类型名称")
		}
	}
	if s.ExprString(params[0].Type) != "context.Context" {
		log.Fatalln(opt.name + "第一个参数必须为context")
	}
	for _, param := range params {
		if len(param.Names) > 1 {
			log.Fatalln(opt.name + "不支持省略类型")
		}
	}
}

func processList(s *pkg.Source, list *ast.Field) (opt options) {
	src := s.Src
	fset := s.Fset
	lines := strings.Split(src, "\n")
	opt = options{Args: s.GetDef(_interfaceName), UseRedis: false, importPackages: s.Packages(list)}
	opt.name = list.Names[0].Name
	opt.KeyMethod = "key" + opt.name
	opt.ExpireCode = "key" + opt.name + "Expire"
	opt.ExpireNullCode = "key" + opt.name + "NullExpire"
	opt.CasVersionCode = "key" + opt.name + "Version"

	ce := fset.Position(list.Pos()).Line - 2
	cs := ce
	cp := regexp.MustCompile(`\s+//.+`)
	for cp.Match([]byte(lines[cs-1])) {
		cs--
	}

	rp := regexp.MustCompile(`//\s*(\w+\s)*rc:.+`)
	optStart := cs
	for optStart <= ce && !rp.Match([]byte(lines[optStart])) {
		optStart++
	}

	// get comment
	if cs < optStart {
		comment := lines[optStart-1]
		opt.Comment = pkg.RegexpReplace(`\s+//(?P<name>.+)`, comment, "$name")
		opt.Comment = strings.TrimSpace(opt.Comment)
	}
	// get options
	getOptions(&opt, lines[optStart:ce+1])
	// get type from prefix
	params := list.Type.(*ast.FuncType).Params.List
	getTypeFromPrefix(&opt, params, s)
	// get template
	if len(params) == 1 {
		opt.template = _noneTpl
	} else if (len(params) == 2) && (opt.RCType == _typeSet || opt.RCType == _typeAdd) {
		if _, ok := params[1].Type.(*ast.MapType); ok {
			opt.template = _multiTpl
		} else {
			opt.template = _noneTpl
		}
	} else {
		if _, ok := params[1].Type.(*ast.ArrayType); ok {
			opt.template = _multiTpl
		} else if _, ok := params[1].Type.(*ast.MapType); ok {
			opt.template = _multiTpl
		} else {
			opt.template = _singleTpl
		}
	}
	// extra args
	if len(params) > _ex {
		args := []string{""}
		allArgs := []string{""}
		pos := 2
		if (opt.RCType == _typeAdd) || (opt.RCType == _typeSet) {
			if opt.template == _singleTpl {
				pos = 3
			}
		}
		for _, pa := range params[pos:] {
			paType := s.ExprString(pa.Type)
			if len(pa.Names) == 0 {
				args = append(args, paType)
				allArgs = append(allArgs, paType)
				continue
			}
			var names []string
			for _, name := range pa.Names {
				names = append(names, name.Name)
			}
			allArgs = append(allArgs, strings.Join(names, ",")+" "+paType)
			args = append(args, strings.Join(names, ","))
		}
		if len(args) > 1 {
			opt.ExtraArgs = strings.Join(args, ",")
			opt.ExtraArgsType = strings.Join(allArgs, ",")
		}
	}
	results := list.Type.(*ast.FuncType).Results.List
	getKeyValueType(&opt, params, results, s)
	return
}

func getKeyValueType(opt *options, params, results []*ast.Field, s *pkg.Source) {
	// check
	if s.ExprString(results[len(results)-1].Type) != "error" {
		log.Fatalln("最后返回值参数需为error")
	}
	for _, res := range results {
		if len(res.Names) > 1 {
			log.Fatalln(opt.name + "返回值不支持省略类型")
		}
	}
	if opt.RCType == _typeGet {
		if len(results) != _pn {
			log.Fatalln("参数个数不对")
		}
	}
	// get key type and value type
	if (opt.RCType == _typeAdd) || (opt.RCType == _typeSet) {
		if opt.template == _multiTpl {
			p, ok := params[1].Type.(*ast.MapType)
			if !ok {
				log.Fatalf("%s: 参数类型错误 批量设置数据时类型需为map类型\n", opt.name)
			}
			opt.keyType = s.ExprString(p.Key)
			opt.ValueType = s.ExprString(p.Value)
		} else if opt.template == _singleTpl {
			opt.keyType = s.ExprString(params[1].Type)
			opt.ValueType = s.ExprString(params[2].Type)
		} else {
			opt.ValueType = s.ExprString(params[1].Type)
		}
	}
	if opt.RCType == _typeGet {
		if opt.template == _multiTpl {
			if p, ok := results[0].Type.(*ast.MapType); ok {
				opt.keyType = s.ExprString(p.Key)
				opt.ValueType = s.ExprString(p.Value)
			} else {
				log.Fatalf("%s: 返回值类型错误 批量获取数据时返回值需为map类型\n", opt.name)
			}
		} else if opt.template == _singleTpl {
			opt.keyType = s.ExprString(params[1].Type)
			opt.ValueType = s.ExprString(results[0].Type)
		} else {
			opt.ValueType = s.ExprString(results[0].Type)
		}
	}
	if opt.RCType == _typeDel {
		if opt.template == _multiTpl {
			p, ok := params[1].Type.(*ast.ArrayType)
			if !ok {
				log.Fatalf("%s: 类型错误 参数需为[]类型\n", opt.name)
			}
			opt.keyType = s.ExprString(p.Elt)
		} else if opt.template == _singleTpl {
			opt.keyType = s.ExprString(params[1].Type)
		}
	}
	for _, t := range simpleTypes {
		if t == opt.ValueType {
			opt.SimpleValue = true
			opt.GetSimpleValue = true
			opt.ConvertValue2Bytes = convertValue2Bytes(t)
			opt.ConvertBytes2Value = convertBytes2Value(t)
			break
		}
	}
	if opt.ValueType == stringType {
		opt.LenType = true
	} else {
		for _, t := range lenTypes {
			if strings.HasPrefix(opt.ValueType, t) {
				opt.LenType = true
				break
			}
		}
	}
	if opt.SimpleValue && (opt.ValueType == "[]byte" || opt.ValueType == stringType) {
		opt.GetSimpleValue = false
		opt.GetDirectValue = true
	}
	if opt.RCType == _typeGet && opt.template == _multiTpl {
		opt.UseRedis = false
	}
	if strings.HasPrefix(opt.ValueType, "*") {
		opt.PointType = true
		opt.OriginValueType = strings.Replace(opt.ValueType, "*", "", 1)
	} else {
		opt.OriginValueType = opt.ValueType
	}
	if *encode != "" {
		var flags []string
		for _, f := range strings.Split(*encode, "|") {
			switch f {
			case "json":
				flags = append(flags, "json")
			case "pb":
				flags = append(flags, "pb")
			default:
				log.Fatalf("%s: encode类型无效\n", opt.name)
			}
		}
		opt.Encode = strings.Join(flags, " | ")
	} else {
		if opt.SimpleValue {
			opt.Encode = "raw"
		} else {
			opt.Encode = "json"
		}
	}
	if strings.Contains(opt.Encode, "json") {
		opt.importPackages = append(opt.importPackages, `"encoding/json"`)
	}
}

func parse(s *pkg.Source) (opts []*options) {
	c := s.F.Scope.Lookup(_interfaceName)
	if (c == nil) || (c.Kind != ast.Typ) {
		log.Fatalln("无法找到缓存声明")
	}
	lists := c.Decl.(*ast.TypeSpec).Type.(*ast.InterfaceType).Methods.List
	for _, list := range lists {
		opt := processList(s, list)
		opt.Check()
		opts = append(opts, &opt)
	}
	return
}

func (option *options) Check() {
	var valid bool
	for _, x := range rcValidTypes {
		if x == option.RCType {
			valid = true
			break
		}
	}
	if !valid {
		log.Fatalf("%s: 类型错误 不支持%s类型\n", option.name, option.RCType)
	}
	if (option.RCType != _typeDel) &&
		!option.SimpleValue &&
		!strings.Contains(option.ValueType, "*") &&
		!strings.Contains(option.ValueType, "[]") &&
		!strings.Contains(option.ValueType, "map") {
		log.Fatalf("%s: 值类型只能为基本类型/slice/map/指针类型\n", option.name)
	}
}

func genHeader(opts []*options) (src string) {
	option := options{PkgName: os.Getenv("GOPACKAGE"), UseRedis: false}
	var packages []string
	packagesMap := map[string]bool{`"context"`: true}
	for _, opt := range opts {
		if opt.CasCode != "" && opt.RCType == _typeSet {
			opt.importPackages = append(opt.importPackages, `"fmt"`)
		} else if opt.CasCode != "" {
			opt.importPackages = append(opt.importPackages, `"time"`)
		}
		if len(opt.importPackages) > 0 {
			for _, p := range opt.importPackages {
				if !packagesMap[p] {
					packages = append(packages, p)
					packagesMap[p] = true
				}
			}
		}
		if opt.Args != "" {
			option.Args = opt.Args
		}
		if opt.UseRedis {
			option.UseRedis = true
		}
		if opt.SimpleValue && !opt.GetDirectValue {
			option.UseStrConv = true
		}
		if opt.EnableBatch {
			option.EnableBatch = true
		}
	}
	option.ImportPackage = strings.Join(packages, "\n")
	src = _headerTemplate
	t := template.Must(template.New("header").Parse(src))
	var buffer bytes.Buffer
	err := t.Execute(&buffer, option)
	if err != nil {
		log.Fatalf("execute template: %s", err)
	}
	// Format the output.
	src = strings.Replace(buffer.String(), "\t", "", -1)
	src = regexp.MustCompile("\n+").ReplaceAllString(src, "\n")
	src = strings.Replace(src, "NEWLINE", "", -1)
	src = strings.Replace(src, "ARGS", option.Args, -1)
	return
}

func getNewTemplate(option *options) (src string) {
	if option.template == _multiTpl {
		switch option.RCType {
		case _typeGet:
			src = _multiGetTemplate
		case _typeSet:
			src = _multiSetTemplate
		case _typeDel:
			src = _multiDelTemplate
		case _typeAdd:
			src = _multiAddTemplate
		}
	} else if option.template == _singleTpl {
		switch option.RCType {
		case _typeGet:
			src = _singleGetTemplate
		case _typeSet:
			src = _singleSetTemplate
		case _typeDel:
			src = _singleDelTemplate
		case _typeAdd:
			src = _singleAddTemplate
		}
	} else {
		switch option.RCType {
		case _typeGet:
			src = _noneGetTemplate
		case _typeSet:
			src = _noneSetTemplate
		case _typeDel:
			src = _noneDelTemplate
		case _typeAdd:
			src = _noneAddTemplate
		}
	}
	return
}

func genBody(opts []*options) (res string) {
	for _, option := range opts {
		src := getNewTemplate(option)
		src = strings.Replace(src, "KEY", option.keyType, -1)
		src = strings.Replace(src, "NAME", option.name, -1)
		src = strings.Replace(src, "VALUE", option.ValueType, -1)
		src = strings.Replace(src, "GROUPSIZE", strconv.Itoa(option.GroupSize), -1)
		src = strings.Replace(src, "MAXGROUP", strconv.Itoa(option.MaxGroup), -1)
		if option.EnableNullCode {
			option.CheckNullCode = strings.Replace(option.CheckNullCode, "$", "val", -1)
		}
		t := template.Must(template.New("cache").Parse(src))
		var buffer bytes.Buffer
		err := t.Execute(&buffer, option)
		if err != nil {
			log.Fatalf("execute template: %s", err)
		}
		// Format the output.
		src = strings.Replace(buffer.String(), "\t", "", -1)
		src = regexp.MustCompile("\n+").ReplaceAllString(src, "\n")
		res = res + "\n" + src
	}
	return
}

// RcCmd represents the rc command.
var RcCmd = &cobra.Command{
	Use:   "rc",
	Short: "redis 缓存代码生成",
	Long:  "redis 缓存代码生成",
	Run:   run,
}

func run(_ *cobra.Command, _ []string) {
	log.SetFlags(0)
	defer func() {
		if err := recover(); err != nil {
			buf := make([]byte, 64*1024)
			buf = buf[:runtime.Stack(buf, false)]
			log.Fatalf("程序解析失败, err: %+v stack: %s", err, buf)
		}
	}()
	options := parse(pkg.NewSource(pkg.SourceText()))
	header := genHeader(options)
	body := genBody(options)
	code := pkg.FormatCode(header + "\n" + body)
	// Write to file.
	dir := filepath.Dir(".")
	outputName := filepath.Join(dir, "dao_rc.go")
	err := os.WriteFile(outputName, []byte(code), 0o644)
	if err != nil {
		log.Fatalf("写入文件失败: %s", err)
	}
	log.Println("dao_rc.go: 生成成功")
}

func convertValue2Bytes(t string) string {
	switch t {
	case "int", "int8", "int16", "int32", "int64":
		return "[]byte(strconv.FormatInt(int64(val), 10))"
	case "uint", "uint8", "uint16", "uint32", "uint64":
		return "[]byte(strconv.FormatUInt(val, 10))"
	case "bool":
		return "[]byte(strconv.FormatBool(val))"
	case "float32":
		return "[]byte(strconv.FormatFloat(val, 'E', -1, 32))"
	case "float64":
		return "[]byte(strconv.FormatFloat(val, 'E', -1, 64))"
	case stringType:
		return "[]byte(val)"
	case "[]byte":
		return "val"
	}
	return ""
}

func convertBytes2Value(t string) string {
	switch t {
	case "int", "int8", "int16", "int32", "int64":
		return "strconv.ParseInt(v, 10, 64)"
	case "uint", "uint8", "uint16", "uint32", "uint64":
		return "strconv.ParseUInt(v, 10, 64)"
	case "bool":
		return "strconv.ParseBool(v)"
	case "float32":
		return "float32(strconv.ParseFloat(v, 32))"
	case "float64":
		return "strconv.ParseFloat(v, 64)"
	}
	return ""
}
