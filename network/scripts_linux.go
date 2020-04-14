package network

import "github.com/phayes/permbits"

const scriptSuffix = ""

const tincUpTxt = `#!/bin/sh
ip addr add {{.Subnet}} dev $INTERFACE
ip link set dev $INTERFACE up
`

const tincDownText = `#!/bin/sh
ip addr del {{.Subnet}} dev $INTERFACE
ip link set dev $INTERFACE down
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
