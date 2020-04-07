package network

type Config struct {
	Name      string
	Port      uint16
	Interface string
	ConnectTo []string
}

type Address struct {
	Host string
	Port uint16
}

type Node struct {
	Name      string
	Subnet    string
	Port      uint16
	Address   []Address
	PublicKey string
}
