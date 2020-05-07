package network

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

type Address struct {
	Host string `json:"host"`
	Port uint16 `json:"port,omitempty"`
}

type Upgrade struct {
	Port    uint16    `json:"port,omitempty"`
	Address []Address `json:"address,omitempty"`
	Device  string    `json:"device,omitempty"`
}

type Node struct {
	Name      string    `json:"name"`
	Subnet    string    `json:"subnet"`
	Port      uint16    `json:"port"`
	Address   []Address `json:"address,omitempty"`
	PublicKey string    `json:"publicKey"`
	Version   int       `json:"version"`
}
