package bts

import (
	"bytes"
	"flag"
	"fmt"
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
	// arguments
	singleFlight  = flag.Bool("singleflight", false, "enable singleflight")
	nullCache     = flag.String("null_cache", "", "null cache")
	checkNullCode = flag.String("bts_check_null_code", "", "check null code")
	cacheErr      = flag.String("cache_err", "continue", "cache err to continue or break")
	batchSize     = flag.Int("bts_batch", 0, "batch size")
	batchErr      = flag.String("bts_batch_err", "break", "batch err to continue or break")
	maxGroup      = flag.Int("bts_max_group", 0, "max group size")
	sync          = flag.Bool("sync", false, "add cache in sync way.")
	paging        = flag.Bool("paging", false, "use paging in single template")
	ignores       = flag.String("ignores", "", "ignore params")
	customMethod  = flag.String("custom_method", "", "自定义方法名 |分隔: 缓存|回源|增加缓存")
	structName    = flag.String("bts_struct_name", "repo", "bts struct name")

	numberTypes = []string{
		"int", "int8", "int16", "int32", "int64",
		"float32", "float64",
		"uint", "uint8", "uint16", "uint32", "uint64",
	}
	simpleTypes = []string{
		"int", "int8", "int16", "int32", "int64",
		"float32", "float64",
		"uint", "uint8", "uint16", "uint32", "uint64",
		"bool", "string", "[]byte",
	}
	optionNames = []string{
		"singleflight", "null_cache", "bts_check_null_code", "bts_batch",
		"bts_max_group", "sync", "paging", "ignores", "bts_batch_err",
		"custom_method", "cache_err", "bts_struct_name",
	}
	optionNamesMap = map[string]bool{}
	interfaceName  string
)

const (
	_multiTpl  = 1
	_singleTpl = 2
	_noneTpl   = 3
	_ex        = 2
	_pn        = 2
	_isn       = 3
	stringType = "string"
)

func resetFlag() {
	*singleFlight = false
	*nullCache = ""
	*checkNullCode = ""
	*batchSize = 0
	*maxGroup = 0
	*sync = false
	*paging = false
	*batchErr = "break"
	*cacheErr = "continue"
	*ignores = ""
	*customMethod = ""
	*structName = "repo"
}

// options options
type options struct {
	name               string
	keyType            string
	valueType          string
	cacheFunc          string
	rawFunc            string
	addCacheFunc       string
	template           int
	SimpleValue        bool
	NumberValue        bool
	GoValue            bool
	ZeroValue          string
	ImportPackage      string
	importPackages     []string
	UseRedis           bool
	Args               string
	PkgName            string
	EnableSingleFlight bool
	NullCache          string
	EnableNullCache    bool
	GroupSize          int
	MaxGroup           int
	EnableBatch        bool
	BatchErrBreak      bool
	Sync               bool
	CheckNullCode      string
	ExtraArgsType      string
	ExtraArgs          string
	ExtraCacheArgs     string
	ExtraRawArgs       string
	ExtraAddCacheArgs  string
	EnablePaging       bool
	Comment            string
	CustomMethod       string
	IDName             string
	CacheErrContinue   bool
	StructName         string
	hasDec             bool
	UseBTS             bool
}

func getOptions(opt *options, comments []string) {
	os.Args = []string{os.Args[0]}
	for _, comment := range comments {
		if regexp.MustCompile(`\s+//\s*(\w+\s)*bts:.+`).Match([]byte(comment)) {
			args := strings.Split(pkg.RegexpReplace(
				`//\s*(\w+\s)*bts:(?P<arg>.+)`, comment, "$arg"), " ",
			)
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
			opt.hasDec = true
		}
	}
	resetFlag()
	flag.Parse()
	opt.EnableSingleFlight = *singleFlight
	opt.NullCache = *nullCache
	opt.EnablePaging = *paging
	opt.EnableNullCache = *nullCache != ""
	opt.EnableBatch = (*batchSize != 0) && (*maxGroup != 0)
	opt.BatchErrBreak = *batchErr == "break"
	opt.Sync = *sync
	opt.CheckNullCode = *checkNullCode
	opt.GroupSize = *batchSize
	opt.MaxGroup = *maxGroup
	opt.CustomMethod = *customMethod
	opt.CacheErrContinue = *cacheErr == "continue"
	opt.StructName = *structName
}

func processList(s *pkg.Source, list *ast.Field) (opt options) {
	fset := s.Fset
	src := s.Src
	lines := strings.Split(src, "\n")
	opt = options{name: list.Names[0].Name, Args: s.GetDef(interfaceName), importPackages: s.Packages(list)}

	ce := fset.Position(list.Pos()).Line - 2
	cs := ce
	cp := regexp.MustCompile(`\s+//.+`)
	for cp.Match([]byte(lines[cs-1])) {
		cs--
	}

	bp := regexp.MustCompile(`//\s*(\w+\s)*bts:.+`)
	optStart := cs
	for optStart <= ce && !bp.Match([]byte(lines[optStart])) {
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
	if !opt.hasDec {
		log.Printf("%s: 无声明 忽略此方法\n", opt.name)
		return
	}
	// get func
	getFunc(s, list, &opt)
	// get k v from results
	results := list.Type.(*ast.FuncType).Results.List
	if len(results) != _pn {
		log.Fatalln(opt.name + ": 参数个数不对")
	}
	if s.ExprString(results[1].Type) != "error" {
		log.Fatalln(opt.name + ": 最后返回值参数需为error")
	}
	if opt.template == _multiTpl {
		p, ok := results[0].Type.(*ast.MapType)
		if !ok {
			log.Fatalln(opt.name + ": 批量获取方法 返回值类型需为map类型")
		}
		opt.keyType = s.ExprString(p.Key)
		opt.valueType = s.ExprString(p.Value)
	} else {
		opt.valueType = s.ExprString(results[0].Type)
	}
	for _, t := range numberTypes {
		if t == opt.valueType {
			opt.NumberValue = true
			break
		}
	}
	opt.ZeroValue = "nil"
	for _, t := range simpleTypes {
		if t == opt.valueType {
			opt.SimpleValue = true
			opt.ZeroValue = zeroValue(t)
			break
		}
	}
	if !opt.SimpleValue {
		for _, t := range []string{"[]", "map"} {
			if strings.HasPrefix(opt.valueType, t) {
				opt.GoValue = true
				break
			}
		}
	}
	upperName := strings.ToUpper(opt.name[0:1]) + opt.name[1:]
	opt.cacheFunc = fmt.Sprintf("repo.Cache%s", upperName)
	opt.rawFunc = fmt.Sprintf("repo.Raw%s", upperName)
	opt.addCacheFunc = fmt.Sprintf("repo.AddCache%s", upperName)
	if opt.CustomMethod != "" {
		arrs := strings.Split(opt.CustomMethod, "|")
		if len(arrs) > 0 && arrs[0] != "" {
			opt.cacheFunc = arrs[0]
		}
		if len(arrs) > 1 && arrs[1] != "" {
			opt.rawFunc = arrs[1]
		}
		if len(arrs) > 2 && arrs[2] != "" {
			opt.addCacheFunc = arrs[2]
		}
	}
	return
}

func getFunc(s *pkg.Source, list *ast.Field, opt *options) {
	params := list.Type.(*ast.FuncType).Params.List
	if len(params) == 0 {
		log.Fatalln(opt.name + "参数不足")
	}
	for _, p := range params {
		if len(p.Names) > 1 {
			log.Fatalln(opt.name + "不支持省略类型 请写全声明中的字段类型名称")
		}
	}
	if s.ExprString(params[0].Type) != "context.Context" {
		log.Fatalln("第一个参数必须为context")
	}
	if len(params) == 1 {
		opt.template = _noneTpl
	} else {
		opt.IDName = params[1].Names[0].Name
		if _, ok := params[1].Type.(*ast.ArrayType); ok {
			opt.template = _multiTpl
		} else {
			opt.template = _singleTpl
			// get key
			opt.keyType = s.ExprString(params[1].Type)
		}
	}
	if len(params) > _ex {
		var args []string
		var allArgs []string
		for _, pa := range params[2:] {
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
			args = append(args, names...)
		}
		opt.ExtraArgs = strings.Join(args, ",")
		opt.ExtraArgsType = strings.Join(allArgs, ",")
		argsMap := make(map[string]bool)
		for _, arg := range args {
			argsMap[arg] = true
		}
		ignoreCache := make(map[string]bool)
		ignoreRaw := make(map[string]bool)
		ignoreAddCache := make(map[string]bool)
		ignoreArray := [3]map[string]bool{ignoreCache, ignoreRaw, ignoreAddCache}
		if *ignores != "" {
			is := strings.Split(*ignores, "|")
			if len(is) > _isn {
				log.Fatalln("ignores参数错误")
			}
			for i := range is {
				if len(is) > i {
					for _, s := range strings.Split(is[i], ",") {
						ignoreArray[i][s] = true
					}
				}
			}
		}
		var as []string
		for _, arg := range args {
			if !ignoreCache[arg] {
				as = append(as, arg)
			}
		}
		opt.ExtraCacheArgs = strings.Join(as, ",")
		as = []string{}
		for _, arg := range args {
			if !ignoreRaw[arg] {
				as = append(as, arg)
			}
		}
		opt.ExtraRawArgs = strings.Join(as, ",")
		as = []string{}
		for _, arg := range args {
			if !ignoreAddCache[arg] {
				as = append(as, arg)
			}
		}
		opt.ExtraAddCacheArgs = strings.Join(as, ",")
		if opt.ExtraAddCacheArgs != "" {
			opt.ExtraAddCacheArgs = "," + opt.ExtraAddCacheArgs
		}
		if opt.ExtraRawArgs != "" {
			opt.ExtraRawArgs = "," + opt.ExtraRawArgs
		}
		if opt.ExtraCacheArgs != "" {
			opt.ExtraCacheArgs = "," + opt.ExtraCacheArgs
		}
		if opt.ExtraArgs != "" {
			opt.ExtraArgs = "," + opt.ExtraArgs
		}
		if opt.ExtraArgsType != "" {
			opt.ExtraArgsType = "," + opt.ExtraArgsType
		}
	}
}

// parse parse options
func parse(s *pkg.Source) (opts []*options) {
	var c *ast.Object
	for _, name := range []string{"_bts", "Dao"} {
		c = s.F.Scope.Lookup(name)
		if (c == nil) || (c.Kind != ast.Typ) {
			c = nil
			continue
		}
		interfaceName = name
		break
	}
	if c == nil || c.Decl == nil {
		log.Fatalln("无法找到缓存声明")
	}

	ts := c.Decl.(*ast.TypeSpec)
	if ts.Type == nil {
		log.Fatalln("无法找到缓存声明")
	}
	it := ts.Type.(*ast.InterfaceType)
	if it.Methods == nil {
		log.Fatalln("无法找到缓存声明")
	}
	lists := it.Methods.List
	for _, list := range lists {
		opt := processList(s, list)
		if opt.hasDec {
			opt.Check()
			opts = append(opts, &opt)
		}
	}
	return
}

func (option *options) Check() {
	if !option.SimpleValue &&
		!strings.Contains(option.valueType, "*") &&
		!strings.Contains(option.valueType, "[]") &&
		!strings.Contains(option.valueType, "map") {
		log.Fatalf("%s: 值类型只能为基本类型/slice/map/指针类型\n", option.name)
	}
	if option.EnableSingleFlight && option.EnableBatch {
		log.Fatalf("%s: 单飞和批量获取不能同时开启\n", option.name)
	}
	if option.template != _singleTpl && option.EnablePaging {
		log.Fatalf("%s: 分页只能用在单key模板中\n", option.name)
	}
	if option.SimpleValue && !option.EnableNullCache {
		if !((option.template == _multiTpl) && option.NumberValue) {
			log.Fatalf("%s: 值为基本类型时需开启空缓存 防止缓存零值穿透\n", option.name)
		}
	}
	if option.EnableNullCache {
		if !option.SimpleValue && option.CheckNullCode == "" {
			log.Fatalf("%s: 缺少-check_null_code参数\n", option.name)
		}
		if option.SimpleValue && option.NullCache == option.ZeroValue {
			log.Fatalf("%s: %s 不能作为空缓存值 \n", option.name, option.NullCache)
		}
		if strings.Contains(option.CheckNullCode, "len") &&
			strings.Contains(strings.Replace(option.CheckNullCode, " ", "", -1), "==0") {
			// -bts_check_null_code=len($)==0 这种无效
			log.Fatalf("%s: -bts_check_null_code=%s 错误 会有无意义的赋值\n", option.name, option.CheckNullCode)
		}
	}
}

func genHeader(opts []*options) (src string) {
	option := options{PkgName: os.Getenv("GOPACKAGE")}
	option.UseBTS = interfaceName == "_bts"
	var sfCount int
	var packages, sfInit []string
	packagesMap := map[string]bool{`"context"`: true}
	if option.template != _multiTpl {
		option.UseRedis = true
		packagesMap[`"github.com/redis/go-redis/v9"`] = true
	}
	for _, opt := range opts {
		if opt.EnableSingleFlight {
			option.EnableSingleFlight = true
			sfCount++
		}
		if opt.EnableBatch {
			option.EnableBatch = true
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
	}
	option.ImportPackage = strings.Join(packages, "\n")
	for i := 0; i < sfCount; i++ {
		sfInit = append(sfInit, "{}")
	}
	src = _headerTemplate
	src = strings.Replace(src, "SFCOUNT", strconv.Itoa(sfCount), -1)
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
	src = strings.Replace(src, "SFINIT", strings.Join(sfInit, ","), -1)
	return
}

func genBody(opts []*options) (res string) {
	sfnum := -1
	for _, option := range opts {
		var nullCodeVar, src string
		if option.template == _multiTpl {
			src = _multiTemplate
			nullCodeVar = "v"
		} else if option.template == _singleTpl {
			src = _singleTemplate
			nullCodeVar = "res"
		} else {
			src = _noneTemplate
			nullCodeVar = "res"
		}
		if option.template != _noneTpl {
			src = strings.Replace(src, "KEY", option.keyType, -1)
		}
		if option.CheckNullCode != "" {
			option.CheckNullCode = strings.Replace(option.CheckNullCode, "$", nullCodeVar, -1)
		}
		if option.EnableSingleFlight {
			sfnum++
		}
		src = strings.Replace(src, "NAME", option.name, -1)
		src = strings.Replace(src, "VALUE", option.valueType, -1)
		src = strings.Replace(src, "ADDCACHEFUNC", option.addCacheFunc, -1)
		src = strings.Replace(src, "CACHEFUNC", option.cacheFunc, -1)
		src = strings.Replace(src, "RAWFUNC", option.rawFunc, -1)
		src = strings.Replace(src, "GROUPSIZE", strconv.Itoa(option.GroupSize), -1)
		src = strings.Replace(src, "MAXGROUP", strconv.Itoa(option.MaxGroup), -1)
		src = strings.Replace(src, "SFNUM", strconv.Itoa(sfnum), -1)
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

func zeroValue(t string) string {
	switch t {
	case "bool":
		return "false"
	case stringType:
		return "\"\""
	case "[]byte":
		return "nil"
	default:
		return "0"
	}
}

func init() {
	for _, name := range optionNames {
		optionNamesMap[name] = true
	}
}

// BtsCmd represents the bts command.
var BtsCmd = &cobra.Command{
	Use:   "bts",
	Short: "缓存回源逻辑代码生成器",
	Long:  "缓存回源逻辑代码生成器",
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
	outputName := filepath.Join(dir, "dao_bts.go")
	err := os.WriteFile(outputName, []byte(code), 0o644)
	if err != nil {
		log.Fatalf("写入文件失败: %s", err)
	}
	log.Println("dao_bts.go: 生成成功")
}
