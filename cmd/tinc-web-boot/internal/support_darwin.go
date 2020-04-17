package internal

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
)

func DetectTincBinary(possibleBinary string) (string, error) {
	if v, err := os.Stat(possibleBinary); err == nil && !v.IsDir() {
		return possibleBinary, nil
	}
	if bin, err := exec.LookPath(possibleBinary); err == nil {
		return bin, nil
	}
	// look int homebrew path
	root := "/usr/local/Cellar/tinc"
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.Name() == "tincd" && !info.IsDir() {
			possibleBinary = path
			return os.ErrExist
		}
		return nil
	})
	if err == os.ErrExist {
		err = nil
	} else if err == nil {
		err = os.ErrNotExist
	}
	return possibleBinary, err
}

func Preload(ctx context.Context) error {
	cmd := exec.CommandContext(ctx, "kextload", "/Library/Extensions/tun.kext", "/Library/Extensions/tap.kext")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
