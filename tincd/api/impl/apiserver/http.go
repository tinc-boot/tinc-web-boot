package apiserver

import (
	"context"
	"github.com/reddec/jsonrpc2"
	"net"
	"net/http"
	"time"
	"tinc-web-boot/tincd/api"
)

func RunHTTP(global context.Context, network, binding string, handler api.API) error {
	listener, err := net.Listen(network, binding)
	if err != nil {
		return err
	}
	var router jsonrpc2.Router
	RegisterAPI(&router, handler)
	server := http.Server{
		Handler: jsonrpc2.HandlerRest(&router),
	}
	ctx, cancel := context.WithCancel(global)
	defer cancel()
	go func() {
		defer listener.Close()
		<-ctx.Done()
		tm, c := context.WithTimeout(context.Background(), 1*time.Second)
		defer c()
		_ = server.Shutdown(tm)
	}()

	return server.Serve(listener)
}
