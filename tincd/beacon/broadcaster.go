package beacon

import (
	"context"
	"fmt"
	"golang.org/x/net/ipv4"
	"log"
	"net"
	"time"
)

const (
	DefaultKeepAlive = 15 * time.Second
)

func Run(ctx context.Context, interfaceName, groupAddress, beacon string) (<-chan Beacon, error) {
	brc := Broadcaster{
		Interface: interfaceName,
		Group:     groupAddress,
		Beacon:    []byte(beacon),
		Context:   ctx,
	}
	return brc.Run()
}

type Broadcaster struct {
	Interface  string          // network interface name
	Group      string          // multicast group address with port
	Beacon     []byte          // beacon to broadcast every interval
	Interval   time.Duration   // (optional, default 15s) interval between beacons
	BufferSize int             // (optional, default 8192) size for buffer for incoming beacons
	Context    context.Context // (optional, default Background) custom context for interruption
}

// Listen UDP multicast address for beacons and send own every interval
func (cfg Broadcaster) Run() (<-chan Beacon, error) {
	group, err := net.ResolveUDPAddr("udp", cfg.Group)
	if err != nil {
		return nil, err
	}

	iface, err := net.InterfaceByName(cfg.Interface)
	if err != nil {
		return nil, err
	}

	addrs, err := iface.Addrs()
	if err != nil {
		return nil, err
	}

	var ip net.IP
	for _, addr := range addrs {
		if v, ok := addr.(*net.IPNet); ok && v.IP.To4() != nil {
			ip = v.IP
			break
		}
	}
	if ip == nil {
		return nil, fmt.Errorf("no IPv4 address on interface")
	}
	log.Println("[TRACE]", "binding broadcaster on", ip.String())

	senderConn, err := net.ListenPacket("udp4", ip.String()+":0")
	if err != nil {
		return nil, err
	}
	senderPacket := ipv4.NewPacketConn(senderConn)

	listener, err := net.ListenUDP("udp4", group)
	if err != nil {
		return nil, err
	}

	packet := ipv4.NewPacketConn(listener)
	err = packet.JoinGroup(iface, group)
	if err != nil {
		listener.Close()
		senderConn.Close()
		return nil, err
	}

	err = senderPacket.SetMulticastLoopback(true)
	if err != nil {
		log.Println("[WARN]", "failed set loopback multicast:", err)
	}

	out := make(chan Beacon, 1)
	go func() {
		defer listener.Close()
		defer senderConn.Close()
		defer close(out)
		cfg.run(packet, senderPacket, group, out)
	}()

	return out, nil
}

func (cfg Broadcaster) run(listener, sender *ipv4.PacketConn, groupAddr *net.UDPAddr, out chan<- Beacon) {
	done := make(chan struct{})
	ctx, cancel := context.WithCancel(cfg.context())
	go func() {
		defer close(done)
		timer := time.NewTicker(cfg.interval())
		defer timer.Stop()
	LOOP:
		for {
			log.Println("[TRACE]", "sending beacon to:", groupAddr, "payload:", string(cfg.Beacon))
			_, _ = sender.WriteTo(cfg.Beacon, nil, groupAddr)
			select {
			case <-ctx.Done():
				break LOOP
			case <-timer.C:
			}
		}
		<-ctx.Done()
		_ = listener.Close()
	}()

	var buffer = make([]byte, cfg.bufferSize())
	log.Println("[TRACE]", "buffer size:", len(buffer))
LOOP:
	for {
		n, _, src, err := listener.ReadFrom(buffer)
		if err != nil {
			break
		}
		if n == 0 {
			continue
		}
		log.Println("[TRACE]", "beacon from:", src, "payload:", string(buffer[:n]))
		cp := make([]byte, n)
		copy(cp, buffer[:n])
		select {
		case out <- Beacon{
			Addr: src,
			Data: cp,
		}:
		case <-ctx.Done():
			break LOOP
		}
	}
	cancel()
	<-done
}

func (cfg Broadcaster) context() context.Context {
	if cfg.Context == nil {
		return context.Background()
	}
	return cfg.Context
}

func (cfg Broadcaster) bufferSize() int {
	if cfg.BufferSize <= 0 {
		return 8192
	}
	return cfg.BufferSize
}

func (cfg Broadcaster) interval() time.Duration {
	if cfg.Interval <= 0 {
		return DefaultKeepAlive
	}
	return cfg.Interval
}
