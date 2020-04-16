package internal

import (
	"context"
	"github.com/pkg/browser"
	"log"
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
			return cmd.Start()
		}
	}
	return browser.OpenURL(url)
}
