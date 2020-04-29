package beacon

import (
	"context"
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

	listener, err := net.ListenMulticastUDP("udp", iface, group)
	if err != nil {
		return nil, err
	}

	out := make(chan Beacon, 1)
	go func() {
		defer listener.Close()
		defer close(out)
		cfg.run(listener, group, out)
	}()

	return out, nil
}

func (cfg Broadcaster) run(listener *net.UDPConn, groupAddr *net.UDPAddr, out chan<- Beacon) {
	done := make(chan struct{})
	ctx, cancel := context.WithCancel(cfg.context())
	go func() {
		defer close(done)
		timer := time.NewTicker(cfg.Interval)
		defer timer.Stop()
	LOOP:
		for {
			_, _ = listener.WriteToUDP(cfg.Beacon, groupAddr)
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
LOOP:
	for {
		n, src, err := listener.ReadFrom(buffer[:])
		if err != nil {
			break
		}
		if n == 0 {
			continue
		}
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
