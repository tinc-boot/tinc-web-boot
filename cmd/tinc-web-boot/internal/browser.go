//+build !windows

package internal

import (
	"log"
	"os"
	"os/exec"
	"os/user"
	"strconv"
	"syscall"
)

var chromes = []string{
	"chrome",
	"google-chrome",
	"google-chrome-stable",
	"chromium",
}

func useOwner(cmd *exec.Cmd) {
	suser := os.Getenv("SUDO_USER")
	if suser == "" {
		return
	}
	info, err := user.Lookup(suser)
	if err != nil {
		return
	}
	uid, err := strconv.ParseUint(info.Uid, 10, 32)
	if err != nil {
		return
	}
	gid, err := strconv.ParseUint(info.Gid, 10, 32)
	if err != nil {
		return
	}
	log.Println("browser will be launched as for user", suser)
	cmd.Env = append(os.Environ(), "USER="+suser, "HOME="+info.HomeDir)
	cmd.SysProcAttr = &syscall.SysProcAttr{}
	cmd.SysProcAttr.Credential = &syscall.Credential{Uid: uint32(uid), Gid: uint32(gid), NoSetGroups: true}
}
