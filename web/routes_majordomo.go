package web

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/reddec/jsonrpc2"
	"tinc-web-boot/network"
	"tinc-web-boot/tincd"
	"tinc-web-boot/web/internal"
	"tinc-web-boot/web/shared"
)

func ExportMajordomo(router gin.IRouter, pool *tincd.Tincd, code string, allowedNetworks ...string) {
	var jsonRouter jsonrpc2.Router
	internal.RegisterTincWebMajordomo(&jsonRouter, NewMajordomo(pool, code, allowedNetworks))
	router.POST("/", gin.WrapH(jsonrpc2.HandlerRest(&jsonRouter)))
}

func NewMajordomo(pool *tincd.Tincd, code string, allowedNetworks []string) *majordomoImpl {
	al := make(map[string]bool)
	for _, a := range allowedNetworks {
		al[a] = true
	}
	return &majordomoImpl{
		pool:    pool,
		allowed: al,
		code:    code,
	}
}

type majordomoImpl struct {
	pool    *tincd.Tincd
	allowed map[string]bool
	code    string
}

func (srv *majordomoImpl) Join(network, code string, self *network.Node) (*shared.Sharing, error) {
	if !srv.allowed[network] {
		return nil, fmt.Errorf("unknown or un-exported network %s", network)
	}
	if code != srv.code {
		return nil, fmt.Errorf("invalid code")
	}
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
