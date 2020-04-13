//

// +build !linux,!darwin

package utils

import "os/exec"

func SetCmdAttrs(cmd *exec.Cmd) {}
