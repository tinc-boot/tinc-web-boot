package tincd

import (
	"context"
	"github.com/gin-gonic/gin"
	"log"
	"net"
	"net/http"
	"strconv"
	"time"
	"tinc-web-boot/network"
)

const (
	retryInterval     = 1 * time.Second
	nodesListInterval = 15 * time.Second
)

func runAPI(ctx context.Context, ntw *network.Network) {
	config, err := ntw.Read()
	if err != nil {
		log.Println(ntw.Name(), ": read config", err)
		return
	}
	bindingHost := config.IP

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
	router := setupRoutes(ntw, config)
	go func() {
		<-ctx.Done()
		listener.Close()
	}()
	_ = router.RunListener(listener)
}

func setupRoutes(ntw *network.Network, config *network.Config) *gin.Engine {
	router := gin.Default()
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
