package serve

import (
	"bytes"
	"fmt"
	"reflect"
	"regexp"
	"strings"
	"text/template"
)

func WriteJS(n *endpointNode, s *server, indent string, write func(...string)) {
	if n.Parent == nil {
		write(renderTemplate(jsPreamble, map[string]any{
			"ApiPath": s.ApiPath,
			"JSFunc":  s.JSFunc,
		}))
	}
	if n.Key != "" {
		validateIdentifier(n.Key)
		write(indent, n.Key)

		// if top-level key output =, else :
		if n.Parent == nil || n.Parent.Parent == nil {
			write(" = ")
		} else {
			write(": ")
		}
	}
	if n.Endpoint != nil {
		write(prefixLines(renderTemplate(jsFuncTemplate, map[string]any{
			"Path":   s.ApiPath + n.Endpoint.Path,
			"JSFunc": s.JSFunc,
			"Args":   Map(n.Endpoint.Args, varName),
		}), indent))
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
			WriteJS(child, s, nextIndent, write)
		}

		if n.Parent != nil {
			write("\n", indent, "}")
		}
	}
}

func varName(t reflect.Type) string {
	t = baseType(t)
	suffix := ""
	for t.Kind() == reflect.Slice {
		suffix += "_arr"
		t = t.Elem()
	}
	return t.Name() + suffix
}

func validateIdentifier(identifier string) {
	if !regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]*$`).MatchString(identifier) {
		panic(fmt.Errorf("invalid identifier: %s", identifier))
	}
}

func prefixLines(text, prefix string) string {
	return strings.Join(strings.Split(text, "\n"), prefix+"\n")
}

func renderTemplate(tpl *template.Template, context any) string {
	buf := bytes.Buffer{}
	PanicOnErr(tpl.Execute(&buf, context))
	return buf.String()
}

var jsPreamble = GetOrPanic(template.New("jsPreamble").Parse(strings.TrimSpace(`
async function {{ .JSFunc }}(path, args) {
  try {
    console.debug('fetching:', path, args)
    const response = await fetch(path, {
      method: 'POST',
      headers: {
        'Accept': 'application/json',
        'Content-Type': 'application/json'
      },
      body: JSON.stringify({Args: args})
    });
    if (!response?.ok) {
      message = await response.text()
      console.error('request for ' + path + ' failed due to: ' + message)
      throw new Error(message)
    }
    const content = await response.json()
    console.debug('got response for', path, args, ':', content)
    return content
  } catch (error) {
    console.log('error fetching', path, 'with args', args, ':', error)
    throw error
  }
}`)))

var jsFuncTemplate = GetOrPanic(template.New("jsFunc").Parse(strings.TrimSpace(`
async function({{ block "args" .Args }}{{- range $j, $p := . }}{{ if $j }}, {{ end }}v{{ $j }}_{{ $p }}{{ end }}{{ end }}){return await {{ .JSFunc }}('{{ .Path }}', [{{template "args" .Args }}])}`)))
