package web

import (
	"tinc-web-boot/network"
	"tinc-web-boot/tincd"
)

type Network struct {
	Name    string          `json:"name"`
	Running bool            `json:"running"`
	Config  *network.Config `json:"config,omitempty"` // only for specific request
}

type Peer struct {
	Name          string        `json:"name"`
	Online        bool          `json:"online"`
	Status        *tincd.Peer   `json:"status,omitempty"`
	Configuration *network.Node `json:"config,omitempty"`
}
