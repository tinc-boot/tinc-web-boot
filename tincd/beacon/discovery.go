package beacon

import (
	"context"
	"sync"
	"time"
)

type Action int

const (
	Discovered Action = 1
	Updated    Action = 2
	Removed    Action = 3
)

type Update struct {
	Address string
	Action  Action
}

func Discovery(ctx context.Context, beacons <-chan Beacon, keepAlive time.Duration) <-chan Update {
	out := make(chan Update, 1)
	go func() {
		defer close(out)
		discovery(ctx, beacons, out, keepAlive)
	}()

	return out
}

func discovery(ctx context.Context, beacons <-chan Beacon, out chan<- Update, keepAlive time.Duration) {
	var cache = make(map[string]time.Time)
	ticker := time.NewTicker(keepAlive / 2)

	defer ticker.Stop()
LOOP:
	for {
		var updates []Update

		select {
		case <-ctx.Done():
			break LOOP
		case <-ticker.C:
			now := time.Now()
			for addr, at := range cache {
				if now.Sub(at) > keepAlive {
					updates = append(updates, Update{
						Address: addr,
						Action:  Removed,
					})
				}
			}
			for _, upd := range updates {
				delete(cache, upd.Address)
			}
		case beacon, ok := <-beacons:
			if !ok {
				break LOOP
			}
			addr := beacon.Addr.String()
			_, has := cache[addr]
			cache[addr] = time.Now()
			action := Discovered
			if has {
				action = Updated
			}
			updates = append(updates, Update{
				Address: addr,
				Action:  action,
			})
		}

		for _, upd := range updates {
			select {
			case out <- upd:
			case <-ctx.Done():
				break LOOP
			}
		}
	}
}

// Handler for new peer. Spawns in a new go-routine.
// Context will be closed when peer disappeared or when source stream closed
// It's guaranteed that only one go-routing will be active per one address
type PeerHandler func(ctx context.Context, address string)

func Peers(updates <-chan Update, handler PeerHandler) {
	type peer struct {
		Cancel func()
		Done   chan struct{}
	}
	ctx, cancel := context.WithCancel(context.Background())

	var peers = make(map[string]*peer)
	var wg sync.WaitGroup

	for update := range updates {
		if update.Action == Discovered {
			child, stop := context.WithCancel(ctx)
			done := make(chan struct{})

			wg.Add(1)
			go func(ctx context.Context, address string) {
				defer wg.Done()
				defer close(done)
				handler(ctx, address)
			}(child, update.Address)

			peers[update.Address] = &peer{
				Cancel: stop,
				Done:   done,
			}

		} else if p, ok := peers[update.Address]; ok && update.Action == Removed {
			p.Cancel()
			<-p.Done
			delete(peers, update.Address)
		}
	}
	cancel()
	wg.Wait()
}
