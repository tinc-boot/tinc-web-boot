package internal

import (
	"context"
	"github.com/pkg/browser"
	"os/exec"
)

var chromes = []string{
	"chrome.exe",
	"google-chrome.exe",
	"google-chrome-stable.exe",
	"chromium.exe",
}

func useOwner(cmd *exec.Cmd) {

}

func runGeneral(ctx context.Context, url string) error {
	return browser.OpenURL(url)
}
