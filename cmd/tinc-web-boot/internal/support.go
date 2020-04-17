// +build !windows,!darwin

package internal

import (
	"context"
	"os/exec"
)

func DetectTincBinary(possibleBinary string) (string, error) {
	return exec.LookPath(possibleBinary)
}

func Preload(ctx context.Context) error {
	return nil
}
