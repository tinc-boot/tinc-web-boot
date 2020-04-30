package network

const scriptSuffix = ".bat"

const tincUpTxt = `
netsh interface ipv4 set address name=%INTERFACE% static {{.IP}}/{{.Mask}} store=persistent
`

const tincDownText = ``

func postProcessScript(filename string) error { return nil }

func ApplyOwnerOfSudoUser(filename string) error { return nil }
