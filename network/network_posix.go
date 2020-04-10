// +build darwin linux

package network

import (
	"context"
)

func (network *Network) postConfigure(ctx context.Context, config *Config, tincBin string) error {
	return nil
}
