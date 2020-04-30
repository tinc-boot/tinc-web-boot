package network

import (
	"bytes"
	"text/template"
)

var (
	tincUpTpl   = template.Must(template.New("").Parse(tincUpTxt))
	tincDownTpl = template.Must(template.New("").Parse(tincDownText))
)

func tincUp(config *Config) string {
	return mustRender(tincUpTpl, config)
}

func tincDown(config *Config) string {
	return mustRender(tincDownTpl, config)
}

func mustRender(tpl *template.Template, params interface{}) string {
	var out bytes.Buffer
	err := tpl.Execute(&out, params)
	if err != nil {
		panic(err)
	}
	return out.String()
}
