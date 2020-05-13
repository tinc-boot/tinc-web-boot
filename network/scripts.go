package network

import (
	"bytes"
	"text/template"
)

var (
	tincUpTpl   = template.Must(template.New("").Parse(tincUpTxt))
	tincDownTpl = template.Must(template.New("").Parse(tincDownText))
)

type params struct {
	Node   *Node
	Config *Config
}

func tincUp(config *Config, node *Node) string {
	return mustRender(tincUpTpl, params{node, config})
}

func tincDown(config *Config, node *Node) string {
	return mustRender(tincDownTpl, params{node, config})
}

func mustRender(tpl *template.Template, params interface{}) string {
	var out bytes.Buffer
	err := tpl.Execute(&out, params)
	if err != nil {
		panic(err)
	}
	return out.String()
}
