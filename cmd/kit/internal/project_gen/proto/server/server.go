package server

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/emicklei/proto"
	"github.com/spf13/cobra"

	"github.com/tianping526/kit/cmd/kit/internal/project_gen/proto/common"
)

const stringType = "string"

// CmdServer the service command.
var CmdServer = &cobra.Command{
	Use:   "server",
	Short: "Generate the proto Server implementations",
	Long:  "Generate Server implementations. Example: kit proto server xip.event.interface.v1 [proto dir] [code dir]",
	Run:   run,
}

func run(_ *cobra.Command, args []string) {
	if len(args) == 0 {
		fmt.Println("need to specify service name.  Example: xip.event.interface.v1 [proto dir] [code dir]")
	}
	input := args[0]
	re := regexp.MustCompile(`^((\w|_)+\.){3}v\d+$`)
	mr := re.Match([]byte(input))
	if !mr {
		fmt.Println("service name err. Example: xip.event.interface.v1")
	}
	parts := strings.Split(input, ".")

	protoDir := "xapis/api"
	codeDir := "app"
	if len(args) > 1 { //nolint:mnd
		protoDir = args[1]
	}
	if len(args) > 2 { //nolint:mnd
		codeDir = args[2]
	}

	servicePath := filepath.Join(common.ModName(), codeDir, strings.Join(parts[:len(parts)-1], "/"))
	codePath := filepath.Join(codeDir, strings.Join(parts[:len(parts)-1], "/"))
	pbPath := filepath.Join(common.ModName(), protoDir, strings.Join(parts, "/"))

	protoPath := filepath.Join(
		protoDir, strings.Join(parts, "/"),
		common.SnakeString(strings.Join(parts, "_")))
	protoPath = protoPath + ".proto"

	reader, err := os.Open(protoPath)
	if err != nil {
		log.Fatal(err)
	}
	defer func(reader *os.File) {
		_ = reader.Close()
	}(reader)

	parser := proto.NewParser(reader)
	definition, err := parser.Parse()
	if err != nil {
		log.Fatal(err)
	}

	var (
		pkg      string
		version  string
		services []*Service
		messages = make(map[string]*Message)
		name     = strings.Join(parts[:len(parts)-1], ".")
	)
	proto.Walk(definition,
		proto.WithOption(func(o *proto.Option) {
			if o.Name == "go_package" {
				sParts := strings.Split(o.Constant.Source, ";")
				pkg = sParts[0]
				version = sParts[1]
			}
		}),
		proto.WithService(func(s *proto.Service) {
			cs := &Service{
				Package:             pkg,
				Service:             s.Name,
				Version:             version,
				ServicePath:         servicePath,
				PbPath:              pbPath,
				Name:                name,
				ServiceName:         getServiceName(s.Name),
				ServiceType:         strings.ToUpper(parts[2][:1]) + parts[2][1:],
				ServiceInternalName: getServiceInternalName(getServiceName(s.Name)),
			}
			for _, e := range s.Elements {
				r, ok := e.(*proto.RPC)
				if !ok {
					continue
				}
				cs.Methods = append(cs.Methods, &Method{
					Service: s.Name, Name: r.Name, Request: r.RequestType,
					Reply: r.ReturnsType, Type: getMethodType(r.StreamsRequest, r.StreamsReturns),
				})
			}
			services = append(services, cs)
		}),
		proto.WithMessage(func(pm *proto.Message) {
			msg := Message{
				Name: pm.Name,
			}
			for _, e := range pm.Elements {
				pf, ok := e.(*proto.NormalField)
				if !ok {
					continue
				}
				field := Field{
					Name:     common.CamelString(pf.Field.Name),
					Type:     pf.Field.Type,
					Repeated: pf.Repeated,
					Optional: pf.Optional,
				}
				if len(field.Name) > msg.FieldMaxLen {
					msg.FieldMaxLen = len(field.Name)
				}
				if pf.Field.Comment != nil {
					field.Comment = pf.Field.Comment.Message()
				}
				if pf.Field.Type == stringType {
					for _, opt := range pf.Field.Options {
						if opt.Name == "(validate.rules).string" {
							if opt.AggregatedConstants[0].Name == "max_bytes" {
								field.LenIfStr = opt.AggregatedConstants[0].Literal.Source
							}
						}
					}
				}
				msg.Field = append(msg.Field, &field)
			}
			messages[pm.Name] = &msg
			parent, ok := pm.Parent.(*proto.Message)
			for ok {
				messages[pm.Name].Parent = append(messages[pm.Name].Parent, parent.Name)
				parent, ok = parent.Parent.(*proto.Message)
			}
		}),
	)
	const empty = "google.protobuf.Empty"
	for _, s := range services {
		msg := map[string]map[string]string{
			"dep":    make(map[string]string),
			"param":  make(map[string]string),
			"import": make(map[string]string),
		}
		for _, method := range s.Methods {
			if (method.Type == unaryType && (method.Request == empty || method.Reply == empty)) ||
				(method.Type == returnsStreamsType && method.Request == empty) {
				s.GoogleEmpty = true
			}
			if method.Type == twoWayStreamsType || method.Type == requestStreamsType {
				s.UseIO = true
			}
			if method.Type == unaryType {
				s.UseContext = true
			}

			_, ok := msg["param"][method.Request]
			if !ok {
				msg["param"][method.Request] = ""
				findDep(messages, msg, method.Request)
			}
			_, ok = msg["param"][method.Reply]
			if !ok {
				msg["param"][method.Reply] = ""
				findDep(messages, msg, method.Reply)
			}

			method.ReqCopy = messages[method.Request].GenerateReqCopyToParam(
				"\t", "req",
				fmt.Sprintf("\t%s := ", toInternalName(dtoCovertDoName(method.Request))),
				messages,
			)
			method.RepCopy = messages[method.Reply].GenerateResCopyToReply(
				"\t", "do",
				fmt.Sprintf("\t%s := ", toInternalName(method.Reply)),
				s.Version,
				messages,
			)

			method.RepDoName = "do"
			if messages[method.Reply].Field == nil {
				method.RepDoName = "_"
			}
		}

		bld := strings.Builder{}
		entities := make([]string, 0, len(msg["dep"])+len(msg["param"]))
		entObj := make(map[string][]string)
		for m := range msg["dep"] {
			entities = append(entities, m)
			entName := TrimMsgPreSuffix(m)
			_, ok := entObj[entName]
			if ok {
				entObj[entName] = append(entObj[entName], m)
			} else {
				entObj[entName] = []string{m}
			}
		}
		for m := range msg["param"] {
			entities = append(entities, m)
			entName := TrimMsgPreSuffix(m)
			_, ok := entObj[entName]
			if ok {
				entObj[entName] = append(entObj[entName], m)
			} else {
				entObj[entName] = []string{m}
			}
		}
		for _, m := range entities {
			bld.WriteString("\n\n")
			bld.WriteString(messages[m].GenerateStruct("", messages))
		}
		s.DO = bld.String()

		ipt := make([]string, 0, len(msg["import"]))
		for k := range msg["import"] {
			ipt = append(ipt, k)
		}
		s.Imports = ipt

		bld.Reset()
		for e := range entObj {
			bld.WriteString("\n\n")
			bld.WriteString(
				fmt.Sprintf("type %s struct {\n\tent.Schema\n}\n\n",
					e,
				),
			)
			bld.WriteString(
				fmt.Sprintf("func (%s) Annotations() []schema.Annotation {\n",
					e,
				),
			)
			bld.WriteString("\treturn []schema.Annotation{\n\t\tentsql.WithComments(true),\n\t}\n}\n\n")
			bld.WriteString(
				fmt.Sprintf("func (%s) Mixin() []ent.Mixin {\n\treturn []ent.Mixin{\n",
					e,
				),
			)
			bld.WriteString("\t\tIDMixin{},\n\t\tmixin.Time{},\n\t}\n}\n\n")
			bld.WriteString(
				fmt.Sprintf("func (%s) Fields() []ent.Field {\n\treturn []ent.Field{\n",
					e,
				),
			)
			for _, em := range entObj[e] {
				bld.WriteString(messages[em].GenerateEntFields("\t", messages))
			}
			bld.WriteString("\t}\n}\n\n")
			bld.WriteString(
				fmt.Sprintf(
					"func (%s) Edges() []ent.Edge {\n\treturn []ent.Edge{}\n}\n\n",
					e,
				),
			)
			bld.WriteString(
				fmt.Sprintf(
					"func (%s) Indexes() []ent.Index {\n\treturn []ent.Index{}\n}",
					e,
				),
			)
		}
		s.Ent = bld.String()
	}

	genService(codePath, services)
	genServer(codePath, services)
	genBiz(codePath, services)
	genData(codePath, services)
	genConf(codePath, services)
	genTest(codePath, services)
}

func getMethodType(streamsRequest, streamsReturns bool) MethodType {
	if !streamsRequest && !streamsReturns {
		return unaryType
	} else if streamsRequest && streamsReturns {
		return twoWayStreamsType
	} else if streamsRequest {
		return requestStreamsType
	} else if streamsReturns {
		return returnsStreamsType
	}
	return unaryType
}

func findDep(messages map[string]*Message, msg map[string]map[string]string, msgType string) {
	for _, f := range messages[msgType].Field {
		if f.Type == pbTimeStamp {
			msg["import"]["google.golang.org/protobuf/types/known/timestamppb"] = ""
		}
		if f.Type == pbDuration {
			msg["import"]["google.golang.org/protobuf/types/known/durationpb"] = ""
		}
		_, ok := messages[f.Type]
		if ok {
			findDep(messages, msg, f.Type)
			msg["dep"][f.Type] = ""
		}
	}
}

func TrimMsgPreSuffix(name string) string {
	name = strings.TrimPrefix(name, "Create")
	name = strings.TrimPrefix(name, "Delete")
	name = strings.TrimPrefix(name, "Get")
	name = strings.TrimPrefix(name, "List")
	name = strings.TrimPrefix(name, "Update")
	name = strings.TrimSuffix(name, "Request")
	name = strings.TrimSuffix(name, "Reply")
	return name
}
