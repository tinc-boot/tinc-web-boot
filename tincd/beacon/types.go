package beacon

import "net"

type Beacon struct {
	Addr net.Addr
	Data []byte
}
