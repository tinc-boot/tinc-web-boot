package beacon

import (
	"bytes"
	"context"
)

func FilterByContent(ctx context.Context, in <-chan Beacon, content []byte) <-chan Beacon {
	out := make(chan Beacon, 1)
	go func() {
		defer close(out)
		for {
			var beacon Beacon
			select {
			case beacon = <-in:
			case <-ctx.Done():
				return
			}

			if !bytes.Equal(beacon.Data, content) {
				continue
			}

			select {
			case out <- beacon:
			case <-ctx.Done():
				return
			}
		}
	}()
	return out
}
