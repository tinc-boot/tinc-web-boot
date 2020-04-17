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
	// look near executable
	execPath, err := os.Executable()
	if err != nil {
		return "", err
	}
	binary := digAround(filepath.Dir(execPath))
	if binary != "" {
		return binary, nil
	}
	return exec.LookPath(possibleBinary)
}

func digAround(dir string) string {
	var ans string
	_ = filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		if info.Name() == "tincd.exe" {
			ans = path
			return os.ErrExist
		}
		return nil
	})
	return ans
}

func Preload(ctx context.Context) error {
	return nil
}
