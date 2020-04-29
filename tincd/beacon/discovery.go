package beacon

import (
	"context"
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
		case beacon := <-beacons:
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
