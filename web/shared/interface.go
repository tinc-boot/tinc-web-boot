package shared

import (
	"context"
	"github.com/tinc-boot/tincd/network"
	"time"
)

type Network struct {
	Name    string          `json:"name"`
	Running bool            `json:"running"`
	Config  *network.Config `json:"config,omitempty"` // only for specific request
}

type PeerInfo struct {
	Name          string       `json:"name"`
	Online        bool         `json:"online"`
	Configuration network.Node `json:"config"`
}

type Sharing struct {
	Name   string          `json:"name"`
	Subnet string          `json:"subnet"`
	Nodes  []*network.Node `json:"node,omitempty"`
}

// Public Tinc-Web API (json-rpc 2.0)
type TincWeb interface {
	// List of available networks (briefly, without config)
	Networks(ctx context.Context) ([]*Network, error)
	// Detailed network info
	Network(ctx context.Context, name string) (*Network, error)
	// Create new network if not exists
	Create(ctx context.Context, name, subnet string) (*Network, error)
	// Remove network (returns true if network existed)
	Remove(ctx context.Context, network string) (bool, error)
	// Start or re-start network
	Start(ctx context.Context, network string) (*Network, error)
	// Stop network
	Stop(ctx context.Context, network string) (*Network, error)
	// Peers brief list in network  (briefly, without config)
	Peers(ctx context.Context, network string) ([]*PeerInfo, error)
	// Peer detailed info by in the network
	Peer(ctx context.Context, network, name string) (*PeerInfo, error)
	// Import another tinc-web network configuration file.
	// It means let nodes defined in config join to the network.
	// Return created (or used) network with full configuration
	Import(ctx context.Context, sharing Sharing) (*Network, error)
	// Share network and generate configuration file.
	Share(ctx context.Context, network string) (*Sharing, error)
	// Node definition in network (aka - self node)
	Node(ctx context.Context, network string) (*network.Node, error)
	// Upgrade node parameters.
	// In some cases requires restart
	Upgrade(ctx context.Context, network string, update network.Upgrade) (*network.Node, error)
	// Generate Majordomo request for easy-sharing
	Majordomo(ctx context.Context, network string, lifetime time.Duration) (string, error)
	// Join by Majordomo Link
	Join(ctx context.Context, url string, start bool) (*Network, error)
}

type EndpointKind string

const (
	Local  EndpointKind = "local"
	Public EndpointKind = "public"
)

type Endpoint struct {
	Host string       `json:"host"`
	Port uint16       `json:"port"`
	Kind EndpointKind `json:"kind"`
}

type Config struct {
	Binding string `json:"binding"`
}

// Operations with tinc-web-boot related to UI
type TincWebUI interface {
	// Issue and sign token
	IssueAccessToken(ctx context.Context, validDays uint) (string, error)
	// Make desktop notification if system supports it
	Notify(ctx context.Context, title, message string) (bool, error)
	// Endpoints list to access web UI
	Endpoints(ctx context.Context) ([]Endpoint, error)
	// Configuration defined for the instance
	Configuration(ctx context.Context) (*Config, error)
}

// Operations for joining public network
type TincWebMajordomo interface {
	// Join public network if code matched. Will generate error if node subnet not matched
	Join(ctx context.Context, network string, self *network.Node) (*Sharing, error)
}
