package tincwebmajordomo

import (
	"context"
	client "github.com/reddec/jsonrpc2/client"
	"sync/atomic"
	network "tinc-web-boot/network"
	shared "tinc-web-boot/web/shared"
)

func Default() *TincWebMajordomoClient {
	return &TincWebMajordomoClient{BaseURL: "http://127.0.0.1:8686/api/"}
}

type TincWebMajordomoClient struct {
	BaseURL  string
	sequence uint64
}

// Join public network if code matched. Will generate error if node subnet not matched
func (impl *TincWebMajordomoClient) Join(ctx context.Context, network string, self *network.Node) (reply *shared.Sharing, err error) {
	err = client.CallHTTP(ctx, impl.BaseURL, "TincWebMajordomo.Join", atomic.AddUint64(&impl.sequence, 1), &reply, network, self)
	return
}
