package web

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"tinc-web-boot/tincd"
)

func New(pool *tincd.Tincd, dev bool) *gin.Engine {
	wrapper := &api{pool: pool}

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

	api := router.Group("/api")
	api.GET("/networks", wrapper.listNetworks)
	api.POST("/networks", wrapper.createNetwork)

	api.GET("/network/:name", wrapper.getNetwork)
	api.DELETE("/network/:name", wrapper.removeNetwork)
	api.POST("/network/:name/status", wrapper.controlNetwork)
	api.GET("/network/:name/peers", wrapper.listPeers)
	api.GET("/network/:name/peer/:peer", wrapper.getPeer)

	return router
}

type api struct {
	pool *tincd.Tincd
}

func (api *api) listNetworks(gctx *gin.Context) {
	var ans []Network
	for _, ntw := range api.pool.Nets() {
		ans = append(ans, Network{
			Name:    ntw.Definition().Name(),
			Running: ntw.IsRunning(),
		})
	}
	gctx.IndentedJSON(http.StatusOK, ans)
}

func (api *api) getNetwork(gctx *gin.Context) {
	ntw, err := api.pool.Get(gctx.Param("name"))
	if err != nil {
		gctx.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	config, err := ntw.Definition().Read()
	if err != nil {
		gctx.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	gctx.IndentedJSON(http.StatusOK, Network{
		Name:    ntw.Definition().Name(),
		Running: ntw.IsRunning(),
		Config:  config,
	})
}

func (api *api) createNetwork(gctx *gin.Context) {
	var params struct {
		Name string `json:"name"`
	}

	if err := gctx.BindJSON(&params); err != nil {
		return
	}

	ntw, err := api.pool.Create(params.Name)
	if err != nil {
		gctx.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	gctx.IndentedJSON(http.StatusOK, Network{
		Name:    ntw.Definition().Name(),
		Running: ntw.IsRunning(),
	})
}

func (api *api) controlNetwork(gctx *gin.Context) {
	var params struct {
		Start bool `json:"start"`
	}

	if err := gctx.BindJSON(&params); err != nil {
		return
	}

	ntw, err := api.pool.Get(gctx.Param("name"))
	if err != nil {
		gctx.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	if params.Start {
		ntw.Start()
	} else {
		ntw.Stop()
	}
	gctx.IndentedJSON(http.StatusOK, Network{
		Name:    ntw.Definition().Name(),
		Running: ntw.IsRunning(),
	})
}

func (api *api) removeNetwork(gctx *gin.Context) {
	err := api.pool.Remove(gctx.Param("name"))
	if err != nil {
		gctx.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	gctx.AbortWithStatus(http.StatusNoContent)
}

func (api *api) listPeers(gctx *gin.Context) {
	ntw, err := api.pool.Get(gctx.Param("name"))
	if err != nil {
		gctx.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	list, err := ntw.Definition().Nodes()
	if err != nil {
		gctx.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	var ans []Peer
	for _, name := range list {
		info, active := ntw.Peer(name)
		ans = append(ans, Peer{
			Name:   name,
			Online: active,
			Status: info,
		})
	}

	gctx.IndentedJSON(http.StatusOK, ans)
}

func (api *api) getPeer(gctx *gin.Context) {
	ntw, err := api.pool.Get(gctx.Param("name"))
	if err != nil {
		gctx.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	node, err := ntw.Definition().Node(gctx.Param("peer"))
	if err != nil {
		gctx.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	info, active := ntw.Peer(node.Name)
	gctx.IndentedJSON(http.StatusOK, Peer{
		Name:          node.Name,
		Online:        active,
		Status:        info,
		Configuration: node,
	})
}
