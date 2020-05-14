package web

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"github.com/reddec/jsonrpc2"
	"github.com/reddec/struct-view/support/events"
	"log"
	"net"
	"net/http"
	"strings"
	"time"
	"tinc-web-boot/network"
	"tinc-web-boot/support/go/tincwebmajordomo"
	"tinc-web-boot/tincd"
	"tinc-web-boot/web/internal"
	"tinc-web-boot/web/shared"
)

const (
	joinTimeout = 15 * time.Second
)

type Config struct {
	Dev             bool
	AuthorizedOnly  bool
	AuthKey         string
	LocalUIPort     uint16
	PublicAddresses []string
	Binding         string
}

//go:generate go-bindata -pkg web -prefix ui/build/ -fs ui/build/...
func (cfg Config) New(pool *tincd.Tincd) (*gin.Engine, *uiRoutes) {

	router := gin.Default()

	if cfg.Dev {
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
	uiApp := &uiRoutes{
		key:           cfg.AuthKey,
		port:          cfg.LocalUIPort,
		publicAddress: cfg.PublicAddresses,
		pool:          pool,
		config:        shared.Config{Binding: cfg.Binding},
	}

	internal.RegisterTincWeb(&jsonRouter, &api{pool: pool, publicAddress: cfg.PublicAddresses, key: cfg.AuthKey})
	internal.RegisterTincWebUI(&jsonRouter, uiApp)

	streamer := events.NewWebsocketStream()
	pool.Events().Sink(streamer.Feed)

	router.StaticFS("/static", AssetFile())

	var majordomoRouter jsonrpc2.Router
	internal.RegisterTincWebMajordomo(&majordomoRouter, NewMajordomo(pool))

	majordomo := router.Group("/majordomo/:token", cfg.majordomoOnly())
	majordomo.POST("", gin.WrapH(jsonrpc2.HandlerRest(&majordomoRouter)))

	api := router.Group("/api/:token/", cfg.authorizedOnly())

	api.POST("", gin.WrapH(jsonrpc2.HandlerRest(&jsonRouter)))
	api.GET("", gin.WrapH(jsonrpc2.HandlerWS(&jsonRouter)))
	api.GET("events", gin.WrapH(streamer))

	router.GET("/", func(gctx *gin.Context) {
		gctx.Redirect(http.StatusTemporaryRedirect, "/static")
	})
	return router, uiApp
}

func (cfg Config) authorizedOnly() gin.HandlerFunc {
	return func(gctx *gin.Context) {
		host, _, _ := net.SplitHostPort(gctx.Request.RemoteAddr)
		if host == "127.0.0.1" && !cfg.AuthorizedOnly {
			// assume localhost connection are authorized
			gctx.Next()
			return
		}
		token := gctx.Param("token")
		_, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return []byte(cfg.AuthKey), nil
		})
		if err != nil {
			log.Println("[guard]", "check token failed:", err)
			gctx.AbortWithStatus(http.StatusForbidden)
			return
		}
		gctx.Next()
	}
}

func (cfg Config) majordomoOnly() gin.HandlerFunc {
	return func(gctx *gin.Context) {
		token := gctx.Param("token")
		claims, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return []byte(cfg.AuthKey), nil
		})
		if err != nil {
			log.Println("[guard]", "check token failed:", err)
			gctx.AbortWithStatus(http.StatusForbidden)
			return
		}
		mp, ok := claims.Claims.(jwt.MapClaims)
		if !ok {
			log.Println("[guard]", "claims not a map")
			gctx.AbortWithStatus(http.StatusForbidden)
			return
		}
		if v, ok := mp["role"].(string); !ok || v != "majordomo" {
			log.Println("[guard]", "wrong role")
			gctx.AbortWithStatus(http.StatusForbidden)
			return
		}
		gctx.Next()
	}
}

type api struct {
	pool          *tincd.Tincd
	key           string
	publicAddress []string
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

	list, err := ntw.Definition().NodesDefinitions()
	if err != nil {
		return nil, err
	}

	var ans []*shared.PeerInfo

	for _, config := range list {
		ans = append(ans, &shared.PeerInfo{
			Name:          config.Name,
			Online:        ntw.IsActive(config.Name),
			Configuration: config,
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
	return &shared.PeerInfo{
		Name:          node.Name,
		Online:        ntw.IsActive(node.Name),
		Configuration: *node,
	}, nil
}

func (srv *api) Create(name, subnet string) (*shared.Network, error) {
	_, cidr, err := net.ParseCIDR(subnet)
	if err != nil {
		return nil, fmt.Errorf("parse subnet: %w", err)
	}
	ntw, err := srv.pool.Create(name, cidr)
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
	_, cidr, err := net.ParseCIDR(sharing.Subnet)
	if err != nil {
		return nil, fmt.Errorf("parse subnet: %w", err)
	}
	ntw, err := srv.pool.Create(sharing.Name, cidr)
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
	return NewShare(ntw.Definition())
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

func (srv *api) Majordomo(network string, lifetime time.Duration) (string, error) {
	if len(srv.publicAddress) == 0 {
		return "", fmt.Errorf("no public addreses defined")
	}
	ntw, err := srv.pool.Get(network)
	if err != nil {
		return "", err
	}
	self, err := ntw.Definition().Self()
	if err != nil {
		return "", err
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"iat":     time.Now().Add(lifetime),
		"role":    "majordomo",
		"subnet":  self.Subnet,
		"network": network,
	})
	tok, err := token.SignedString([]byte(srv.key))
	if err != nil {
		return "", err
	}

	return "http://" + srv.publicAddress[0] + "/majordomo/" + tok, nil
}

func NewShare(ntw *network.Network) (*shared.Sharing, error) {
	nodeNames, err := ntw.Nodes()
	if err != nil {
		return nil, err
	}
	var ans shared.Sharing
	ans.Name = ntw.Name()

	for _, name := range nodeNames {
		node, err := ntw.Node(name)
		if err != nil {
			return nil, fmt.Errorf("get node %s: %w", name, err)
		}
		ans.Nodes = append(ans.Nodes, node)
		ans.Subnet = node.Subnet
	}

	return &ans, nil
}

func (srv *api) Join(url string, start bool) (*shared.Network, error) {
	parts := strings.Split(url, "/")
	token := parts[len(parts)-1]
	data := strings.Split(token, ".")[1]
	bindata, err := base64.RawStdEncoding.DecodeString(data)
	if err != nil {
		return nil, err
	}

	var share struct {
		Network string `json:"network"`
		Subnet  string `json:"subnet"`
	}

	err = json.Unmarshal(bindata, &share)
	if err != nil {
		return nil, err
	}

	remote := &tincwebmajordomo.TincWebMajordomoClient{BaseURL: url}

	ntw, err := srv.Create(share.Network, share.Subnet)
	if err != nil {
		return nil, err
	}

	self, err := srv.Node(ntw.Name)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), joinTimeout)
	defer cancel()
	sharedNet, err := remote.Join(ctx, share.Network, self)
	if err != nil {
		return nil, err
	}

	info, err := srv.Import(*sharedNet)
	if err != nil {
		return nil, err
	}
	if start {
		return srv.Start(info.Name)
	}
	return info, nil
}
