package main

import (
	"context"
	"github.com/alecthomas/kong"
	"log"
	"os"
	"os/signal"
	"time"
	"tinc-web-boot/cmd/tinc-web-boot/internal"
	"tinc-web-boot/network"
	"tinc-web-boot/tincd"
)

type Main struct {
	APIPort int    `name:"api-port" env:"API_PORT" help:"API port" default:"18655"`
	TincBin string `name:"tinc-bin" env:"TINC_BIN" help:"Custom tinc binary location" default:"tincd"`
	Host    string `name:"host" env:"HOST" help:"Binding host" default:"127.0.0.1"`
	Dir     string `name:"dir" env:"DIR" help:"Directory for config" default:"networks"`
}

func main() {
	var cli Main
	ctx := kong.Parse(&cli)
	var err error
	if ctx.Command() == "" {
		err = cli.Run()
	} else {
		err = ctx.Run(nil)
	}
	ctx.FatalIfErrorf(err)
}

func (m *Main) Run() error {
	binary, err := internal.DetectTincBinary(m.TincBin)
	if err != nil {
		return err
	}
	log.Println("Detected Tinc binary:", binary)

	ctx, closer := context.WithCancel(context.Background())
	go func() {
		c := make(chan os.Signal, 2)
		signal.Notify(c, os.Kill, os.Interrupt)
		for range c {
			closer()
			break
		}
	}()
	defer closer()

	stor := &network.Storage{Root: m.Dir}
	err = stor.Init()
	if err != nil {
		return err
	}

	pool, err := tincd.New(ctx, stor, m.APIPort, binary)
	if err != nil {
		return err
	}
	defer pool.Stop()

	time.Sleep(10 * time.Second)

	return ctx.Err()
}
