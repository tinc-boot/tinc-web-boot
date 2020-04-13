package network

import (
	"context"
)

func (network *Network) postConfigure(ctx context.Context, config *Config, tincBin string) error {
	return nil
}

func (network *Network) beforeConfigure(config *Config) error {
	config.Interface = ""
	return nil
}
