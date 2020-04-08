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
}
