package main

import (
	"github.com/alecthomas/kong"
	"log"
	"tinc-web-boot/cmd/tinc-web-boot/internal"
)

type Main struct {
	APIPort int    `name:"api-port" env:"API_PORT" help:"API port" default:"18655"`
	TincBin string `name:"tinc-bin" env:"TINC_BIN" help:"Custom tinc binary location" default:"tincd"`
	Host    string `name:"host" env:"HOST" help:"Binding host" default:"127.0.0.1"`
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
	return nil
}
