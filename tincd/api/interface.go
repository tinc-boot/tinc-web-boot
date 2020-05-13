package api

import "tinc-web-boot/network"

type API interface {
	// Send self description and get known nodes
	Exchange(self network.Node) ([]network.Node, error)
}
