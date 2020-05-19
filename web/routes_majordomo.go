package web

import (
	"context"
	"fmt"
	"github.com/tinc-boot/tincd/network"

	"tinc-web-boot/pool"
	"tinc-web-boot/web/shared"
)

func NewMajordomo(pool *pool.Pool) *majordomoImpl {
	return &majordomoImpl{
		pool: pool,
	}
}

type majordomoImpl struct {
	pool *pool.Pool
}

func (srv *majordomoImpl) Join(ctx context.Context, network string, self *network.Node) (*shared.Sharing, error) {
	ntw, err := srv.pool.Network(network)
	if err != nil {
		return nil, err
	}
	err = ntw.Put(self)
	if err != nil {
		return nil, fmt.Errorf("import node %s: %w", self.Name, err)
	}
	return NewShare(ntw)
}
