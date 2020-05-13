package tincweb

import (
	"context"
	client "github.com/reddec/jsonrpc2/client"
	"sync/atomic"
	"time"
	network "tinc-web-boot/network"
	shared "tinc-web-boot/web/shared"
)

func Default() *TincWebClient {
	return &TincWebClient{BaseURL: "http://127.0.0.1:8686/api/"}
}

type TincWebClient struct {
	BaseURL  string
	sequence uint64
}

// List of available networks (briefly, without config)
func (impl *TincWebClient) Networks(ctx context.Context) (reply []*shared.Network, err error) {
	err = client.CallHTTP(ctx, impl.BaseURL, "TincWeb.Networks", atomic.AddUint64(&impl.sequence, 1), &reply)
	return
}

// Detailed network info
func (impl *TincWebClient) Network(ctx context.Context, name string) (reply *shared.Network, err error) {
	err = client.CallHTTP(ctx, impl.BaseURL, "TincWeb.Network", atomic.AddUint64(&impl.sequence, 1), &reply, name)
	return
}

// Create new network if not exists
func (impl *TincWebClient) Create(ctx context.Context, name string, subnet string) (reply *shared.Network, err error) {
	err = client.CallHTTP(ctx, impl.BaseURL, "TincWeb.Create", atomic.AddUint64(&impl.sequence, 1), &reply, name, subnet)
	return
}

// Remove network (returns true if network existed)
func (impl *TincWebClient) Remove(ctx context.Context, network string) (reply bool, err error) {
	err = client.CallHTTP(ctx, impl.BaseURL, "TincWeb.Remove", atomic.AddUint64(&impl.sequence, 1), &reply, network)
	return
}

// Start or re-start network
func (impl *TincWebClient) Start(ctx context.Context, network string) (reply *shared.Network, err error) {
	err = client.CallHTTP(ctx, impl.BaseURL, "TincWeb.Start", atomic.AddUint64(&impl.sequence, 1), &reply, network)
	return
}

// Stop network
func (impl *TincWebClient) Stop(ctx context.Context, network string) (reply *shared.Network, err error) {
	err = client.CallHTTP(ctx, impl.BaseURL, "TincWeb.Stop", atomic.AddUint64(&impl.sequence, 1), &reply, network)
	return
}

// Peers brief list in network  (briefly, without config)
func (impl *TincWebClient) Peers(ctx context.Context, network string) (reply []*shared.PeerInfo, err error) {
	err = client.CallHTTP(ctx, impl.BaseURL, "TincWeb.Peers", atomic.AddUint64(&impl.sequence, 1), &reply, network)
	return
}

// Peer detailed info by in the network
func (impl *TincWebClient) Peer(ctx context.Context, network string, name string) (reply *shared.PeerInfo, err error) {
	err = client.CallHTTP(ctx, impl.BaseURL, "TincWeb.Peer", atomic.AddUint64(&impl.sequence, 1), &reply, network, name)
	return
}

/*
Import another tinc-web network configuration file.
It means let nodes defined in config join to the network.
Return created (or used) network with full configuration
*/
func (impl *TincWebClient) Import(ctx context.Context, sharing shared.Sharing) (reply *shared.Network, err error) {
	err = client.CallHTTP(ctx, impl.BaseURL, "TincWeb.Import", atomic.AddUint64(&impl.sequence, 1), &reply, sharing)
	return
}

// Share network and generate configuration file.
func (impl *TincWebClient) Share(ctx context.Context, network string) (reply *shared.Sharing, err error) {
	err = client.CallHTTP(ctx, impl.BaseURL, "TincWeb.Share", atomic.AddUint64(&impl.sequence, 1), &reply, network)
	return
}

// Node definition in network (aka - self node)
func (impl *TincWebClient) Node(ctx context.Context, network string) (reply *network.Node, err error) {
	err = client.CallHTTP(ctx, impl.BaseURL, "TincWeb.Node", atomic.AddUint64(&impl.sequence, 1), &reply, network)
	return
}

/*
Upgrade node parameters.
In some cases requires restart
*/
func (impl *TincWebClient) Upgrade(ctx context.Context, network string, update network.Upgrade) (reply *network.Node, err error) {
	err = client.CallHTTP(ctx, impl.BaseURL, "TincWeb.Upgrade", atomic.AddUint64(&impl.sequence, 1), &reply, network, update)
	return
}

// Generate Majordomo request for easy-sharing
func (impl *TincWebClient) Majordomo(ctx context.Context, network string, lifetime time.Duration) (reply string, err error) {
	err = client.CallHTTP(ctx, impl.BaseURL, "TincWeb.Majordomo", atomic.AddUint64(&impl.sequence, 1), &reply, network, lifetime)
	return
}
