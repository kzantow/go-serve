package serve

import (
	"bytes"
	"fmt"
	"html/template"
	"regexp"
	"strings"
)

type endpointNode struct {
	Key      string
	Parent   *endpointNode
	Endpoint *endpoint
	SubKeys  map[string]*endpointNode
}

func (n *endpointNode) WriteJS(s *server, indent string, write func(...string)) {
	if n.Parent == nil {
		write(renderTemplate(preamble, map[string]any{
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
			"Params": n.Endpoint.Params,
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
			child.WriteJS(s, nextIndent, write)
		}

		if n.Parent != nil {
			write("\n", indent, "}")
		}
	}
}

func (n *endpointNode) Children() []*endpointNode {
	return SortedMapValues(n.SubKeys)
}

func (n *endpointNode) getNode(parts []string) *endpointNode {
	if len(parts) == 0 {
		return n
	}
	part := parts[0]
	if n.SubKeys == nil {
		n.SubKeys = map[string]*endpointNode{}
	}
	node := n.SubKeys[part]
	if node == nil {
		node = &endpointNode{
			Key:    part,
			Parent: n,
		}
		n.SubKeys[part] = node
	}
	return node.getNode(parts[1:])
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

var preamble = GetOrPanic(template.New("preamble").Parse(strings.TrimSpace(`
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
async function({{ block "args" .Params }}{{- range $j, $p := . }}{{ if $j }},{{ end }}v{{ $j }}_{{ $p }}{{ end }}{{ end }}){return await {{ .JSFunc }}('{{ .Path }}',[{{template "args" .Params }}])}`)))
