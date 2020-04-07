package tincd

import (
	"context"
	"github.com/gin-gonic/gin"
	"log"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"
	"tinc-web-boot/network"
)

const (
	retryInterval     = 1 * time.Second
	communicationPort = 1655
)

func runAPI(ctx context.Context, requests chan<- peerReq, network *network.Network) {
	config, err := network.Read()
	if err != nil {
		log.Println(network.Name(), ": read config", err)
		return
	}
	selfNode, err := network.Node(config.Name)
	if err != nil {
		log.Println(network.Name(), ": read self node", err)
		return
	}
	bindingHost := strings.TrimSpace(strings.Split(selfNode.Subnet, "/")[0])

	lc := &net.ListenConfig{}
	var listener net.Listener
	for {

		l, err := lc.Listen(ctx, "tcp", bindingHost+":"+strconv.Itoa(communicationPort))
		if err == nil {
			listener = l
			break
		}
		log.Println(network.Name(), "listen:", err)
		select {
		case <-ctx.Done():
			return
		case <-time.After(retryInterval):
		}
	}
	defer listener.Close()
	router := setupRoutes(ctx, requests, network, config)
	go func() {
		<-ctx.Done()
		listener.Close()
	}()
	_ = router.RunListener(listener)
}

func setupRoutes(ctx context.Context, requests chan<- peerReq, network *network.Network, config *network.Config) *gin.Engine {
	router := gin.Default()
	router.POST("/rpc/watch", func(gctx *gin.Context) {
		var params struct {
			Subnet string `json:"subnet"`
			Node   string `json:"node"`
		}
		if err := gctx.BindJSON(&params); err != nil {
			return
		}
		select {
		case requests <- peerReq{
			Node:   params.Node,
			Subnet: params.Subnet,
			Add:    true,
		}:
		case <-ctx.Done():

		}
		gctx.AbortWithStatus(http.StatusNoContent)
	})

	router.POST("/rpc/forget", func(gctx *gin.Context) {
		var params struct {
			Node string `json:"node"`
		}
		if err := gctx.BindJSON(&params); err != nil {
			return
		}
		select {
		case requests <- peerReq{
			Node: params.Node,
			Add:  false,
		}:
		case <-ctx.Done():

		}
		gctx.AbortWithStatus(http.StatusNoContent)
	})

	router.GET("/", func(gctx *gin.Context) {
		gctx.File(network.NodeFile(config.Name))
	})

	return router
}
