package web

import (
	"github.com/gin-gonic/gin"
	"github.com/reddec/jsonrpc2"
	"tinc-web-boot/tincd"
	"tinc-web-boot/web/internal"
	"tinc-web-boot/web/shared"
)

func New(pool *tincd.Tincd, dev bool) *gin.Engine {

	router := gin.Default()

	if dev {
		router.Use(func(gctx *gin.Context) {
			gctx.Writer.Header().Set("Access-Control-Allow-Origin", "*")
			gctx.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
			gctx.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
			gctx.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

			if gctx.Request.Method == "OPTIONS" {
				gctx.AbortWithStatus(204)
				return
			}

			gctx.Next()
		})
	} else {
		router.Use(func(gctx *gin.Context) {
			gctx.Writer.Header().Set("X-XSS-Protection", "1; mode=block")
			gctx.Writer.Header().Set("X-Frame-Options", "SAMEORIGIN")
			gctx.Writer.Header().Set("X-Content-Type-Options", "nosniff")
			gctx.Next()
		})
	}

	var jsonRouter jsonrpc2.Router
	internal.RegisterTincWeb(&jsonRouter, &api{pool: pool})

	router.POST("/api", gin.WrapH(jsonrpc2.Handler(&jsonRouter)))

	return router
}

type api struct {
	pool *tincd.Tincd
}

func (srv *api) Networks() ([]*shared.Network, error) {
	var ans []*shared.Network
	for _, ntw := range srv.pool.Nets() {
		ans = append(ans, &shared.Network{
			Name:    ntw.Definition().Name(),
			Running: ntw.IsRunning(),
		})
	}
	return ans, nil
}

func (srv *api) Network(name string) (*shared.Network, error) {
	ntw, err := srv.pool.Get(name)
	if err != nil {
		return nil, err
	}
	config, err := ntw.Definition().Read()
	if err != nil {
		return nil, err
	}
	return &shared.Network{
		Name:    ntw.Definition().Name(),
		Running: ntw.IsRunning(),
		Config:  config,
	}, nil
}

func (srv *api) Peers(network string) ([]*shared.PeerInfo, error) {
	ntw, err := srv.pool.Get(network)
	if err != nil {
		return nil, err
	}

	list, err := ntw.Definition().Nodes()
	if err != nil {
		return nil, err
	}

	var ans []*shared.PeerInfo
	for _, name := range list {
		info, active := ntw.Peer(name)
		ans = append(ans, &shared.PeerInfo{
			Name:   name,
			Online: active,
			Status: info,
		})
	}
	return ans, nil
}

func (srv *api) Peer(network, name string) (*shared.PeerInfo, error) {
	ntw, err := srv.pool.Get(network)
	if err != nil {
		return nil, err
	}
	node, err := ntw.Definition().Node(name)
	if err != nil {
		return nil, err
	}
	info, active := ntw.Peer(node.Name)
	return &shared.PeerInfo{
		Name:          node.Name,
		Online:        active,
		Status:        info,
		Configuration: node,
	}, nil
}

func (srv *api) Create(name string) (*shared.Network, error) {
	ntw, err := srv.pool.Create(name)
	if err != nil {
		return nil, err
	}
	return &shared.Network{
		Name:    ntw.Definition().Name(),
		Running: ntw.IsRunning(),
	}, nil
}

func (srv *api) Remove(network string) (bool, error) {
	exists, err := srv.pool.Remove(network)
	return exists, err
}

func (srv *api) Start(network string) (*shared.Network, error) {
	ntw, err := srv.pool.Get(network)
	if err != nil {
		return nil, err
	}
	ntw.Start()
	return &shared.Network{
		Name:    ntw.Definition().Name(),
		Running: ntw.IsRunning(),
	}, nil
}

func (srv *api) Stop(network string) (*shared.Network, error) {
	ntw, err := srv.pool.Get(network)
	if err != nil {
		return nil, err
	}
	ntw.Stop()
	return &shared.Network{
		Name:    ntw.Definition().Name(),
		Running: ntw.IsRunning(),
	}, nil
}
