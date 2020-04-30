package web

import (
	"fmt"
	"tinc-web-boot/network"
	"tinc-web-boot/tincd"
	"tinc-web-boot/web/shared"
)

func NewMajordomo(pool *tincd.Tincd) *majordomoImpl {
	return &majordomoImpl{
		pool: pool,
	}
}

type majordomoImpl struct {
	pool *tincd.Tincd
}

func (srv *majordomoImpl) Join(network string, self *network.Node) (*shared.Sharing, error) {
	ntw, err := srv.pool.Get(network)
	if err != nil {
		return nil, err
	}
	err = ntw.Definition().Put(self)
	if err != nil {
		return nil, fmt.Errorf("import node %s: %w", self.Name, err)
	}
	return NewShare(ntw.Definition())
}
