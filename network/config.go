package network

import (
	"fmt"
	"strconv"
	"strings"
	"tinc-web-boot/config"
)

const (
	CommunicationPort = 1655
)

type Config struct {
	Name       string   `json:"name"`
	Port       uint16   `json:"port"`
	Interface  string   `json:"interface"`
	AutoStart  bool     `json:"autostart"`
	Mode       string   `json:"mode"`
	IP         string   `json:"ip"`
	Mask       int      `json:"mask"`
	DeviceType string   `json:"deviceType,omitempty"`
	Device     string   `json:"device,omitempty"`
	ConnectTo  []string `json:"connectTo,omitempty"`
	Broadcast  string   `json:"broadcast"`
}

type Upgrade struct {
	Port    uint16    `json:"port,omitempty"`
	Address []Address `json:"address,omitempty"`
	Device  string    `json:"device,omitempty"`
}

type Address struct {
	Host string `json:"host"`
	Port uint16 `json:"port,omitempty"`
}

func (addr *Address) String() string {
	if addr.Port != 0 {
		return fmt.Sprintf("%s %v", addr.Host, addr.Port)
	}
	return addr.Host
}

func (addr *Address) Scan(value string) error {
	hp := strings.SplitN(strings.TrimSpace(value), " ", 2)
	addr.Host = hp[0]
	if len(hp) == 1 {
		return nil
	}
	v, err := strconv.ParseUint(hp[1], 10, 16)
	addr.Port = uint16(v)
	return err
}

type Node struct {
	Name      string    `json:"name"`
	Subnet    string    `json:"subnet"`
	Port      uint16    `json:"port"`
	Address   []Address `json:"address,omitempty"`
	PublicKey string    `json:"publicKey" tinc:"RSA PUBLIC KEY,blob"`
	Version   int       `json:"version"`
}

func (cfg *Config) Build() (text []byte, err error) {
	return config.Marshal(cfg)
}

func (cfg *Config) Parse(text []byte) error {
	return config.Unmarshal(text, cfg)
}

func (n *Node) Build() (text []byte, err error) {
	return config.Marshal(n)
}

func (n *Node) Parse(data []byte) error {
	return config.Unmarshal(data, n)
}
