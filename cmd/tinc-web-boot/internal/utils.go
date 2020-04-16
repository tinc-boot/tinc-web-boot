package internal

import (
	"context"
	"log"
	"os"
	"os/exec"
)

func OpenInBrowser(ctx context.Context, url string, app bool) error {
	if app {
		for _, name := range chromes {
			bin, err := exec.LookPath(name)
			if err != nil {
				continue
			}
			log.Println("open", url, "as app")
			cmd := exec.CommandContext(ctx, bin, "--app="+url)
			cmd.Stderr = os.Stderr
			cmd.Stdout = os.Stdout
			useOwner(cmd)
			return cmd.Start()
		}
	}
	return runGeneral(ctx, url)
}
