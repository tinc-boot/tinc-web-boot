package web

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/reddec/jsonrpc2"
	"github.com/reddec/struct-view/support/events"
	"net/http"
	"tinc-web-boot/network"
	"tinc-web-boot/tincd"
	"tinc-web-boot/web/internal"
	"tinc-web-boot/web/shared"
)

//go:generate go-bindata -pkg web -prefix ui/build/ -fs ui/build/...
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

			gctx.Request.Header.Del("Origin")

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

	streamer := events.NewWebsocketStream()
	pool.Events().Sink(streamer.Feed)

	router.POST("/api", gin.WrapH(jsonrpc2.HandlerRest(&jsonRouter)))
	router.GET("/api", gin.WrapH(jsonrpc2.HandlerWS(&jsonRouter)))
	router.GET("/api/events", gin.WrapH(streamer))
	router.StaticFS("/static", AssetFile())
	router.GET("/", func(gctx *gin.Context) {
		gctx.Redirect(http.StatusTemporaryRedirect, "/static")
	})

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

func (srv *api) Import(sharing shared.Sharing) (*shared.Network, error) {
	ntw, err := srv.pool.Create(sharing.Name)
	if err != nil {
		return nil, err
	}

	config, err := ntw.Definition().Read()
	if err != nil {
		return nil, err
	}

	for _, node := range sharing.Nodes {
		err := ntw.Definition().Put(node)
		if err != nil {
			return nil, fmt.Errorf("import node %s: %w", node.Name, err)
		}
	}

	return &shared.Network{
		Name:    ntw.Definition().Name(),
		Running: ntw.IsRunning(),
		Config:  config,
	}, nil
}

func (srv *api) Share(network string) (*shared.Sharing, error) {
	ntw, err := srv.pool.Get(network)
	if err != nil {
		return nil, err
	}
	nodeNames, err := ntw.Definition().Nodes()
	if err != nil {
		return nil, err
	}
	var ans shared.Sharing
	ans.Name = network

	for _, name := range nodeNames {
		node, err := ntw.Definition().Node(name)
		if err != nil {
			return nil, fmt.Errorf("get node %s: %w", name, err)
		}
		ans.Nodes = append(ans.Nodes, node)
	}

	return &ans, nil
}

func (srv *api) Upgrade(network string, update network.Upgrade) (*network.Node, error) {
	ntw, err := srv.pool.Get(network)
	if err != nil {
		return nil, err
	}
	err = ntw.Definition().Upgrade(update)
	if err != nil {
		return nil, err
	}
	return srv.Node(network)
}

func (srv *api) Node(network string) (*network.Node, error) {
	ntw, err := srv.pool.Get(network)
	if err != nil {
		return nil, err
	}
	cfg, err := ntw.Definition().Read()
	if err != nil {
		return nil, err
	}
	return ntw.Definition().Node(cfg.Name)
}
