package network

import (
	"fmt"
	"github.com/phayes/permbits"
	"log"
	"os"
	"os/user"
	"strconv"
)

const scriptSuffix = ""

const tincUpTxt = `#!/bin/sh
ip addr add {{.Node.IP}}/{{.Config.Mask}} dev $INTERFACE
ip link set dev $INTERFACE up
`

const tincDownText = `#!/bin/sh
ip addr del {{.Node.IP}}/{{.Config.Mask}} dev $INTERFACE
ip link set dev $INTERFACE down
`

func postProcessScript(filename string) error {
	if err := ApplyOwnerOfSudoUser(filename); err != nil {
		log.Println("post-process", filename, ":", err)
	}
	stat, err := permbits.Stat(filename)
	if err != nil {
		return err
	}
	stat.SetGroupExecute(true)
	stat.SetOtherExecute(true)
	stat.SetUserExecute(true)
	return permbits.Chmod(filename, stat)
}

func ApplyOwnerOfSudoUser(filename string) error {
	suser := os.Getenv("SUDO_USER")
	if suser == "" {
		return nil
	}
	info, err := user.Lookup(suser)
	if err != nil {
		return fmt.Errorf("lookup %s: %w", suser, err)
	}
	uid, err := strconv.Atoi(info.Uid)
	if err != nil {
		return fmt.Errorf("parse UID %s: %w", suser, err)
	}
	gid, err := strconv.Atoi(info.Gid)
	if err != nil {
		return fmt.Errorf("parse GID %s: %w", suser, err)
	}
	return os.Chown(filename, uid, gid)
}
