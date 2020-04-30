package tincwebmajordomo

import (
	"context"
	client "github.com/reddec/jsonrpc2/client"
	"sync/atomic"
)

//
type Node struct {
	Name      string    `json:"name"`
	Subnet    string    `json:"subnet"`
	Port      uint16    `json:"port"`
	Address   []Address `json:"address"`
	PublicKey string    `json:"publicKey"`
	Version   int       `json:"version"`
}

//
type Address struct {
	Host string `json:"host"`
	Port uint16 `json:"port"`
}

//
type Sharing struct {
	Name   string  `json:"name"`
	Subnet string  `json:"subnet"`
	Nodes  []*Node `json:"node"`
}

func Default() *TincWebMajordomoClient {
	return &TincWebMajordomoClient{BaseURL: "http://127.0.0.1:8686/api/"}
}

type TincWebMajordomoClient struct {
	BaseURL  string
	sequence uint64
}

// Join public network if code matched. Will generate error if node subnet not matched
func (impl *TincWebMajordomoClient) Join(ctx context.Context, network string, self *Node) (reply *Sharing, err error) {
	err = client.CallHTTP(ctx, impl.BaseURL, "TincWebMajordomo.Join", atomic.AddUint64(&impl.sequence, 1), &reply, network, self)
	return
}
