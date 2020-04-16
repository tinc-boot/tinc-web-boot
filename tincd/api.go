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
	nodesListInterval = 15 * time.Second
)

func runAPI(ctx context.Context, requests chan<- peerReq, ntw *network.Network) {
	config, err := ntw.Read()
	if err != nil {
		log.Println(ntw.Name(), ": read config", err)
		return
	}
	selfNode, err := ntw.Node(config.Name)
	if err != nil {
		log.Println(ntw.Name(), ": read self node", err)
		return
	}
	bindingHost := strings.TrimSpace(strings.Split(selfNode.Subnet, "/")[0])

	lc := &net.ListenConfig{}
	var listener net.Listener
	for {

		l, err := lc.Listen(ctx, "tcp", bindingHost+":"+strconv.Itoa(network.CommunicationPort))
		if err == nil {
			listener = l
			break
		}
		log.Println(ntw.Name(), "listen:", err)
		select {
		case <-ctx.Done():
			return
		case <-time.After(retryInterval):
		}
	}
	defer listener.Close()
	router := setupRoutes(ctx, requests, ntw, config)
	go func() {
		<-ctx.Done()
		listener.Close()
	}()
	_ = router.RunListener(listener)
}

func setupRoutes(ctx context.Context, requests chan<- peerReq, ntw *network.Network, config *network.Config) *gin.Engine {
	router := gin.Default()
	router.POST("/rpc/watch", func(gctx *gin.Context) {
		var params struct {
			Subnet string `json:"subnet"`
			Node   string `json:"node"`
		}
		if err := gctx.BindJSON(&params); err != nil {
			return
		}
		if _, _, err := net.ParseCIDR(params.Subnet); err != nil {
			log.Printf("incorrect subnet (%s) found: %v", params.Subnet, err)
			gctx.AbortWithError(http.StatusBadRequest, err)
			return
		} else {
			log.Println("detected new subnet", params.Subnet, "belongs to", params.Node)
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

	router.GET("/rpc/nodes", func(gctx *gin.Context) {
		var nodes nodeList
		names, err := ntw.Nodes()
		if err != nil {
			gctx.AbortWithError(http.StatusInternalServerError, err)
			return
		}

		for _, name := range names {
			node, err := ntw.Node(name)
			if err != nil {
				gctx.AbortWithError(http.StatusInternalServerError, err)
				return
			}

			nodes.Nodes = append(nodes.Nodes, node)
		}

		gctx.IndentedJSON(http.StatusOK, nodes)
	})

	router.GET("/", func(gctx *gin.Context) {
		gctx.File(ntw.NodeFile(config.Name))
	})

	return router
}

type nodeList struct {
	Nodes []*network.Node `json:"nodes"`
}
