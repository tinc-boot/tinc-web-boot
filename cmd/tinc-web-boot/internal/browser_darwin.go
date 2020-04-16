package internal

import (
	"context"
	"os"
	"os/exec"
)

func runGeneral(ctx context.Context, url string) error {
	cmd := exec.CommandContext(ctx, "open", url)
	useOwner(cmd)
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	return cmd.Start()
}
