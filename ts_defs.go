package serve

import (
	"fmt"
	"reflect"
	"slices"
	"strings"
	"text/template"
)

func WriteTypeScriptDefinitions(n *endpointNode, s *server, indent string, write func(...string)) {
	if n.Parent == nil {
		write(renderTemplate(tsPreamble, nil), "\n")
		// output all custom return types
		var output []reflect.Type
		for _, e := range s.EndpointList() {
			for _, returnTyp := range e.Returns {
				WriteTSdType(&output, returnTyp, write)
			}
		}
	}
	if n.Key != "" {
		validateIdentifier(n.Key)

		if n.Parent == nil || n.Parent.Parent == nil {
			write("declare var ", n.Key, ": ", n.Key, "\ninterface ")
		}

		write(indent, n.Key)
	}
	if n.Endpoint != nil {
		write("(")
		for i, arg := range n.Endpoint.Args {
			if i > 0 {
				write(", ")
			}
			write("v", fmt.Sprintf("%v", i), "_", varName(arg), ": ", tsType(arg))
		}
		write("): Promise<")
		switch len(n.Endpoint.Returns) {
		case 0:
		case 1:
			write(tsType(n.Endpoint.Returns[0]))
		default:
			write("[")
			for i, arg := range n.Endpoint.Returns {
				if i > 0 {
					write(", ")
				}
				write("v", fmt.Sprintf("%v", i), "_", varName(arg), ": ", tsType(arg))
			}
			write("]")
		}
		write(">")
	} else {
		if n.Parent != nil {
			write("{")
		}

		nextIndent := indent
		if n.Parent != nil {
			nextIndent = indent + "  "
		}
		for i, child := range n.Children() {
			if n.Parent != nil && i > 0 {
				write(",")
			}
			write("\n")
			WriteTypeScriptDefinitions(child, s, nextIndent, write)
		}

		if n.Parent != nil {
			write("\n", indent, "}")
		}
	}
}

func WriteTSdType(completed *[]reflect.Type, typ reflect.Type, write func(...string)) {
	typ = baseElemType(typ)
	if typ.Kind() != reflect.Struct {
		return
	}
	if slices.Contains(*completed, typ) {
		return
	}
	var discoveredStructs []reflect.Type
	write("interface ", typ.Name(), " {\n")
	for i := 0; i < typ.NumField(); i++ {
		f := typ.Field(i)
		if !f.IsExported() {
			continue
		}
		if f.Tag.Get("json") == "-" {
			continue
		}
		if baseElemType(f.Type).Kind() == reflect.Struct {
			discoveredStructs = append(discoveredStructs, baseElemType(f.Type))
		}
		write("  ", f.Name, "?: ", tsType(f.Type), "\n")
	}
	write("}\n")
	*completed = append(*completed, typ)

	for _, t := range discoveredStructs {
		WriteTSdType(completed, t, write)
	}
}

func baseElemType(t reflect.Type) reflect.Type {
	t = baseType(t)
	for t.Kind() == reflect.Slice {
		t = t.Elem()
	}
	return baseType(t)
}

func tsType(t reflect.Type) string {
	switch t {
	case byteSlice:
		return "base64" // this gets automatically base64 encoded as a string
	}
	suffix := ""
	for t.Kind() == reflect.Slice {
		suffix += "[]"
		t = t.Elem()
	}
	name := t.Name()
	switch t.Kind() {
	case reflect.String:
		name = "string"
	case reflect.Int,
		reflect.Int8,
		reflect.Int16,
		reflect.Int32,
		reflect.Int64,
		reflect.Uint,
		reflect.Uint8,
		reflect.Uint16,
		reflect.Uint32,
		reflect.Uint64,
		reflect.Float32,
		reflect.Float64:
		name = "number"
	default:
	}
	if name == "" {
		name = "any"
	}
	return name + suffix
}

var byteSlice = reflect.TypeOf([]byte{})

var tsPreamble = GetOrPanic(template.New("tsPreamble").Parse(strings.TrimSpace(`
type base64 = string
`)))
