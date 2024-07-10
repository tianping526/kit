package server

import (
	"fmt"
	"strings"

	"github.com/gertd/go-pluralize"

	"github.com/tianping526/kit/cmd/kit/internal/project_gen/proto/common"
)

const (
	pbTimeStamp = "google.protobuf.Timestamp"
	pbDuration  = "google.protobuf.Duration"
)

var plural = pluralize.NewClient()

type Message struct {
	Name        string
	FieldMaxLen int
	Parent      []string
	Field       []*Field
}

type Field struct {
	Name     string
	Type     string
	Comment  string
	Repeated bool
	Optional bool
	LenIfStr string
}

func (m *Message) GenerateStruct(tab string, ctx map[string]*Message) string {
	bld := strings.Builder{}
	bld.WriteString(tab)
	bld.WriteString("type ")
	bld.WriteString(dtoCovertDoName(m.Name))
	bld.WriteString(" struct")
	if len(m.Field) != 0 {
		bld.WriteString(" {\n")
		for _, f := range m.Field {
			bld.WriteString(tab)
			bld.WriteString("\t")
			bld.WriteString(fmt.Sprintf(fmt.Sprintf("%%-%ds", m.FieldMaxLen), idToID(f.Name)))
			bld.WriteString(" ")
			if f.Repeated {
				_, ok := ctx[f.Type]
				if ok {
					bld.WriteString("[]*")
				} else {
					bld.WriteString("[]")
				}
			} else {
				_, ok := ctx[f.Type]
				if f.Optional || ok {
					bld.WriteString("*")
				}
			}
			bld.WriteString(msgTypeToStructType(f.Type))
			bld.WriteString("\n")
		}
		bld.WriteString(tab)
	} else {
		bld.WriteString("{")
	}
	bld.WriteString("}")
	return bld.String()
}

func msgTypeToStructType(t string) string {
	switch t {
	case pbTimeStamp:
		return "*timestamppb.Timestamp"
	case pbDuration:
		return "*durationpb.Duration"
	default:
		return t
	}
}

func (m *Message) GenerateReqCopyToParam(
	tab string,
	src string,
	dest string,
	ctx map[string]*Message,
) string {
	bld := strings.Builder{}
	for _, f := range m.Field {
		internalName := toInternalName(f.Name)
		singularName := toInternalName(plural.Singular(f.Name))
		if f.Repeated {
			_, ok := ctx[f.Type]
			if ok {
				bld.WriteString(tab)
				bld.WriteString(fmt.Sprintf(
					"%s := make([]*biz.%s, 0, len(%s.%s))\n",
					internalName,
					f.Type,
					src,
					f.Name,
				))
				bld.WriteString(tab)
				bld.WriteString(fmt.Sprintf(
					"for _, %s := range %s.%s {\n", singularName, src, f.Name,
				))
				_, ok := ctx[f.Type]
				if ok {
					bld.WriteString(ctx[f.Type].GenerateReqCopyToParam(
						tab+"\t", singularName, fmt.Sprintf("%s\t%s := ", tab, singularName), ctx,
					))
					bld.WriteString("\n")
				}
				bld.WriteString(fmt.Sprintf("%s\t%s = append(%s, %s)\n",
					tab, internalName, internalName, singularName,
				))
				bld.WriteString(tab)
				bld.WriteString("}\n")
			}
		}
	}

	bld.WriteString(dest)
	bld.WriteString("&biz.")
	bld.WriteString(dtoCovertDoName(m.Name))
	bld.WriteString("{")
	if len(m.Field) != 0 {
		bld.WriteString("\n")
		for _, f := range m.Field {
			bld.WriteString(tab)
			bld.WriteString("\t")
			bld.WriteString(fmt.Sprintf(
				fmt.Sprintf("%%-%ds", m.FieldMaxLen+1),
				fmt.Sprintf("%s:", idToID(f.Name))),
			)
			bld.WriteString(" ")
			_, ok := ctx[f.Type]
			if ok {
				bld.WriteString(toInternalName(f.Name))
			} else {
				bld.WriteString(msgTypeToStructType(fmt.Sprintf("%s.", src)))
				bld.WriteString(f.Name)
			}
			bld.WriteString(",\n")
		}
		bld.WriteString(tab)
	}
	bld.WriteString("}")
	return bld.String()
}

func toInternalName(name string) string {
	return strings.ToLower(string(name[0])) + name[1:]
}

func (m *Message) GenerateResCopyToReply(
	tab string,
	src string,
	dest string,
	pkg string,
	ctx map[string]*Message,
) string {
	bld := strings.Builder{}
	for _, f := range m.Field {
		internalName := toInternalName(f.Name)
		singularName := toInternalName(plural.Singular(f.Name))
		if f.Repeated {
			_, ok := ctx[f.Type]
			if ok {
				bld.WriteString(tab)
				bld.WriteString(fmt.Sprintf(
					"%s := make([]*%s.%s, 0, len(%s.%s))\n",
					internalName,
					pkg,
					ctx[f.Type].msgCovertDtoName(),
					src,
					f.Name,
				))
				bld.WriteString(tab)
				bld.WriteString(fmt.Sprintf(
					"for _, %s := range %s.%s {\n", singularName, src, f.Name,
				))
				_, ok := ctx[f.Type]
				if ok {
					bld.WriteString(ctx[f.Type].GenerateResCopyToReply(
						tab+"\t",
						singularName,
						fmt.Sprintf("%s\t%s := ", tab, singularName),
						pkg,
						ctx,
					))
					bld.WriteString("\n")
				}
				bld.WriteString(fmt.Sprintf("%s\t%s = append(%s, %s)\n",
					tab, internalName, internalName, singularName,
				))
				bld.WriteString(tab)
				bld.WriteString("}\n")
			}
		}
	}

	bld.WriteString(dest)
	bld.WriteString(fmt.Sprintf("&%s.", pkg))
	bld.WriteString(m.msgCovertDtoName())
	bld.WriteString("{")
	if len(m.Field) != 0 {
		bld.WriteString("\n")
		for _, f := range m.Field {
			bld.WriteString(tab)
			bld.WriteString("\t")
			bld.WriteString(fmt.Sprintf(
				fmt.Sprintf("%%-%ds", m.FieldMaxLen+1),
				fmt.Sprintf("%s:", f.Name)),
			)
			bld.WriteString(" ")
			_, ok := ctx[f.Type]
			if ok {
				bld.WriteString(toInternalName(f.Name))
			} else {
				bld.WriteString(msgTypeToStructType(fmt.Sprintf("%s.", src)))
				bld.WriteString(idToID(f.Name))
			}
			bld.WriteString(",\n")
		}
		bld.WriteString(tab)
	}
	bld.WriteString("}")
	return bld.String()
}

func (m *Message) GenerateEntFields(tab string, ctx map[string]*Message) string {
	bld := strings.Builder{}
	for _, f := range m.Field {
		bld.WriteString(tab)
		bld.WriteString("\tfield.")
		bld.WriteString(msgCovertEntType(f.Type, ctx))
		bld.WriteString("(\"")
		bld.WriteString(common.SnakeString(f.Name))
		bld.WriteString("\").\n")
		if f.LenIfStr != "" {
			bld.WriteString(tab)
			bld.WriteString("\t\tMaxLen(")
			bld.WriteString(f.LenIfStr)
			bld.WriteString(").\n")
		}
		bld.WriteString(tab)
		bld.WriteString("\t\tComment(\"")
		bld.WriteString(strings.Trim(f.Comment, " "))
		bld.WriteString("\"),\n")
	}
	return bld.String()
}

func dtoCovertDoName(name string) string {
	if strings.HasSuffix(name, "Request") {
		return name[:len(name)-7] + "Param"
	}
	if strings.HasSuffix(name, "Reply") {
		return name[:len(name)-5] + "Res"
	}
	return name
}

func (m *Message) msgCovertDtoName() string {
	dtoName := ""
	if m.Parent != nil {
		for i := len(m.Parent) - 1; i >= 0; i-- {
			dtoName += m.Parent[i] + "_"
		}
	}
	dtoName += m.Name
	return dtoName
}

func msgCovertEntType(Type string, ctx map[string]*Message) string {
	switch Type {
	case pbTimeStamp, pbDuration:
		return "Time"
	default:
		_, ok := ctx[Type]
		if ok {
			return "String"
		}
		return common.CamelString(Type)
	}
}

func idToID(s string) string {
	data := make([]byte, 0, len(s))
	num := len(s) - 1
	for i := 0; i <= num; i++ {
		d := s[i]

		if d == 'I' && num > i && s[i+1] == 'd' {
			if i == 0 || s[i-1] >= 'Z' || s[i-1] <= 'A' {
				if num == i+1 || (s[i-1] >= 'A' && s[i-1] <= 'Z') {
					data = append(data, d)
					i++
					data = append(data, 'D')
					continue
				}
			}
		}
		data = append(data, d)
	}
	return string(data[:])
}
