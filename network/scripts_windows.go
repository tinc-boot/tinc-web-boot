package network

const scriptSuffix = ".bat"

const tincUpTxt = `
netsh interface ipv4 set address name="%INTERFACE%" static {{.Subnet}} store=persistent
`

const tincDownText = ``

const subnetUpText = `
{{.Executable}} subnet add /api-port {{.Port}}
`

const subnetDownText = `
{{.Executable}} subnet remove /api-port {{.Port}}
`

func postProcessScript(filename string) error { return nil }
