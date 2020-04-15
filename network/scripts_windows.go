package network

const scriptSuffix = ".bat"

const tincUpTxt = `
netsh interface ipv4 set address name="%INTERFACE%" static {{.Subnet}} store=persistent
`

const tincDownText = ``

const subnetUpText = `
{{.Executable}} subnet add
`

const subnetDownText = `
{{.Executable}} subnet remove
`

func postProcessScript(filename string) error { return nil }

func ApplyOwnerOfSudoUser(filename string) error { return nil }
