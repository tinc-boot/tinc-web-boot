package network

import "github.com/phayes/permbits"

const scriptSuffix = ""

const tincUpTxt = `#!/bin/sh
ifconfig $INTERFACE {{.Subnet}} {{.IP}}
ifconfig $INTERFACE mtu 1350
route add -net {{.IP}} {{.IP}} 255.255.255.255
`

const tincDownText = `#!/bin/sh
ifconfig $INTERFACE down
`

const subnetUpText = `#!/bin/sh
{{.Executable}} subnet add
`

const subnetDownText = `#!/bin/sh
{{.Executable}} subnet remove
`

func postProcessScript(filename string) error {
	stat, err := permbits.Stat(filename)
	if err != nil {
		return err
	}
	stat.SetGroupExecute(true)
	stat.SetOtherExecute(true)
	stat.SetUserExecute(true)
	return permbits.Chmod(filename, stat)
}
