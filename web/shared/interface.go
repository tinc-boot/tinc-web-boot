package shared

import (
	"tinc-web-boot/network"
	"tinc-web-boot/tincd"
)

type Network struct {
	Name    string          `json:"name"`
	Running bool            `json:"running"`
	Config  *network.Config `json:"config,omitempty"` // only for specific request
}

type PeerInfo struct {
	Name          string        `json:"name"`
	Online        bool          `json:"online"`
	Status        *tincd.Peer   `json:"status,omitempty"`
	Configuration *network.Node `json:"config,omitempty"`
}

type Sharing struct {
	Name  string          `json:"name"`
	Nodes []*network.Node `json:"node,omitempty"`
}

// Public Tinc-Web API (json-rpc 2.0)
type TincWeb interface {
	// List of available networks (briefly, without config)
	Networks() ([]*Network, error)
	// Detailed network info
	Network(name string) (*Network, error)
	// Create new network if not exists
	Create(name string) (*Network, error)
	// Remove network (returns true if network existed)
	Remove(network string) (bool, error)
	// Start or re-start network
	Start(network string) (*Network, error)
	// Stop network
	Stop(network string) (*Network, error)
	// Peers brief list in network  (briefly, without config)
	Peers(network string) ([]*PeerInfo, error)
	// Peer detailed info by in the network
	Peer(network, name string) (*PeerInfo, error)
	// Import another tinc-web network configuration file.
	// It means let nodes defined in config join to the network.
	// Return created (or used) network with full configuration
	Import(sharing Sharing) (*Network, error)
	// Share network and generate configuration file.
	Share(network string) (*Sharing, error)
	// Node definition in network (aka - self node)
	Node(network string) (*network.Node, error)
	// Upgrade node parameters.
	// In some cases requires restart
	Upgrade(network string, update network.Upgrade) (*network.Node, error)
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

// Operations with tinc-web-boot related to UI
type TincWebUI interface {
	// Issue and sign token
	IssueAccessToken(validDays uint) (string, error)
	// Make desktop notification if system supports it
	Notify(title, message string) (bool, error)
	// Endpoints list to access web UI
	Endpoints() ([]Endpoint, error)
}
