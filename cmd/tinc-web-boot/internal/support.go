package internal

import (
	"os/exec"
)

func DetectTincBinary(possibleBinary string) (string, error) {
	return exec.LookPath(possibleBinary)
}
