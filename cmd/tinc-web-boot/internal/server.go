package internal

import (
	"context"
	"log"
	"net/http"
	"time"
)

type HttpServer struct {
	GracefulShutdown time.Duration `name:"graceful-shutdown" env:"GRACEFUL_SHUTDOWN" help:"Interval before server shutdown" default:"15s" json:"graceful_shutdown"`
	Bind             string        `name:"bind" env:"BIND" help:"Address to where bind HTTP server" default:"127.0.0.1:8686" json:"bind"`
	TLS              bool          `name:"tls" env:"TLS" help:"Enable HTTPS serving with TLS" json:"tls"`
	CertFile         string        `name:"cert-file" env:"CERT_FILE" help:"Path to certificate for TLS" default:"server.crt" json:"crt_file"`
	KeyFile          string        `name:"key-file" env:"KEY_FILE" help:"Path to private key for TLS" default:"server.key" json:"key_file"`
}

func (qs *HttpServer) Serve(globalCtx context.Context, handler http.Handler) error {

	server := http.Server{
		Addr:    qs.Bind,
		Handler: handler,
	}

	go func() {
		<-globalCtx.Done()
		ctx, cancel := context.WithTimeout(context.Background(), qs.GracefulShutdown)
		defer cancel()
		server.Shutdown(ctx)
	}()
	log.Println("REST server is on", qs.Bind)
	if qs.TLS {
		return server.ListenAndServeTLS(qs.CertFile, qs.KeyFile)
	}
	return server.ListenAndServe()
}
