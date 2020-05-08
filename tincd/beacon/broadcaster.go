package beacon

import (
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"log"
	"net"
	"time"
)

const (
	DefaultKeepAlive = 15 * time.Second
)

func Run(ctx context.Context, interfaceName, beacon string, port uint16) (<-chan Beacon, error) {
	brc := Broadcaster{
		Interface: interfaceName,
		Beacon:    []byte(beacon),
		Context:   ctx,
		Port:      port,
	}
	return brc.Run()
}

type Broadcaster struct {
	Interface  string          // network interface name
	Port       uint16          // broadcasting port
	Beacon     []byte          // beacon to broadcast every interval
	Interval   time.Duration   // (optional, default 15s) interval between beacons
	BufferSize int             // (optional, default 8192) size for buffer for incoming beacons
	Context    context.Context // (optional, default Background) custom context for interruption
}

// Listen UDP multicast address for beacons and send own every interval
func (cfg Broadcaster) Run() (<-chan Beacon, error) {
	iface, err := net.InterfaceByName(cfg.Interface)
	if err != nil {
		return nil, err
	}

	addrs, err := iface.Addrs()
	if err != nil {
		return nil, err
	}

	var (
		ip            net.IP
		broadcastAddr net.IP
	)
	for _, addr := range addrs {
		if v, ok := addr.(*net.IPNet); ok && v.IP.To4() != nil {
			ip = v.IP
			bcast, err := getBroadcastAddr(v)
			if err != nil {
				return nil, err
			}
			broadcastAddr = bcast
			break
		}
	}
	if ip == nil {
		return nil, fmt.Errorf("no IPv4 address on interface")
	}
	bindingAddr := fmt.Sprintf("%s:%d", ip.String(), cfg.Port)
	log.Println("[TRACE]", "binding broadcaster on", bindingAddr)

	socket, err := net.ListenUDP("udp4", &net.UDPAddr{
		IP:   ip,
		Port: int(cfg.Port),
	})
	if err != nil {
		return nil, err
	}

	out := make(chan Beacon, 1)
	go func() {
		defer socket.Close()
		defer close(out)
		cfg.run(broadcastAddr, socket, out)
	}()

	return out, nil
}

func (cfg Broadcaster) run(broadcast net.IP, socket net.PacketConn, out chan<- Beacon) {
	done := make(chan struct{})
	ctx, cancel := context.WithCancel(cfg.context())
	go func() {
		defer close(done)
		timer := time.NewTicker(cfg.interval())
		defer timer.Stop()
		dest := &net.UDPAddr{IP: broadcast, Port: int(cfg.Port)}
	LOOP:
		for {
			log.Println("[TRACE]", "sending beacon to", dest)
			_, _ = socket.WriteTo(cfg.Beacon, dest)
			select {
			case <-ctx.Done():
				break LOOP
			case <-timer.C:
			}
		}
		<-ctx.Done()
		_ = socket.Close()
	}()

	var buffer = make([]byte, cfg.bufferSize())
	log.Println("[TRACE]", "buffer size:", len(buffer))
LOOP:
	for {
		n, src, err := socket.ReadFrom(buffer)
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

// https://stackoverflow.com/a/36167611/1195316
func getBroadcastAddr(n *net.IPNet) (net.IP, error) { // works when the n is a prefix, otherwise...
	if n.IP.To4() == nil {
		return net.IP{}, errors.New("does not support IPv6 addresses")
	}
	ip := make(net.IP, len(n.IP.To4()))
	binary.BigEndian.PutUint32(ip, binary.BigEndian.Uint32(n.IP.To4())|^binary.BigEndian.Uint32(net.IP(n.Mask).To4()))
	return ip, nil
}
