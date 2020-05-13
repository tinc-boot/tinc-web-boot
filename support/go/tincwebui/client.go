package tincwebui

import (
	"context"
	client "github.com/reddec/jsonrpc2/client"
	"sync/atomic"
	shared "tinc-web-boot/web/shared"
)

func Default() *TincWebUIClient {
	return &TincWebUIClient{BaseURL: "http://127.0.0.1:8686/api/"}
}

type TincWebUIClient struct {
	BaseURL  string
	sequence uint64
}

// Issue and sign token
func (impl *TincWebUIClient) IssueAccessToken(ctx context.Context, validDays uint) (reply string, err error) {
	err = client.CallHTTP(ctx, impl.BaseURL, "TincWebUI.IssueAccessToken", atomic.AddUint64(&impl.sequence, 1), &reply, validDays)
	return
}

// Make desktop notification if system supports it
func (impl *TincWebUIClient) Notify(ctx context.Context, title string, message string) (reply bool, err error) {
	err = client.CallHTTP(ctx, impl.BaseURL, "TincWebUI.Notify", atomic.AddUint64(&impl.sequence, 1), &reply, title, message)
	return
}

// Endpoints list to access web UI
func (impl *TincWebUIClient) Endpoints(ctx context.Context) (reply []shared.Endpoint, err error) {
	err = client.CallHTTP(ctx, impl.BaseURL, "TincWebUI.Endpoints", atomic.AddUint64(&impl.sequence, 1), &reply)
	return
}

// Configuration defined for the instance
func (impl *TincWebUIClient) Configuration(ctx context.Context) (reply *shared.Config, err error) {
	err = client.CallHTTP(ctx, impl.BaseURL, "TincWebUI.Configuration", atomic.AddUint64(&impl.sequence, 1), &reply)
	return
}
