package network

const scriptSuffix = ".bat"

const tincUpTxt = `
netsh interface ipv4 set address name="%INTERFACE%" static {{.Subnet}} store=persistent
`

const tincDownText = ``

const subnetUpText = `
{{.Executable}} subnet add && route add "$SUBNET" {{.Node.IP}}
`

const subnetDownText = `
{{.Executable}} subnet remove && route delete "$SUBNET"
`

func postProcessScript(filename string) error { return nil }

func ApplyOwnerOfSudoUser(filename string) error { return nil }
