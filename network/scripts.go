package network

import (
	"bytes"
	"text/template"
)

var (
	tincUpTpl   = template.Must(template.New("").Parse(tincUpTxt))
	tincDownTpl = template.Must(template.New("").Parse(tincDownText))
)

func tincUp(selfNode *Node) string {
	return mustRender(tincUpTpl, selfNode)
}

func tincDown(selfNode *Node) string {
	return mustRender(tincDownTpl, selfNode)
}

func mustRender(tpl *template.Template, params interface{}) string {
	var out bytes.Buffer
	err := tpl.Execute(&out, params)
	if err != nil {
		panic(err)
	}
	return out.String()
}
