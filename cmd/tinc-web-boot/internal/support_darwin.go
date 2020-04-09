package internal

import (
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
			possibleBinary = filepath.Join(path, info.Name())
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
