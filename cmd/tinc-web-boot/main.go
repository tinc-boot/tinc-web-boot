package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/alecthomas/kong"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"time"
	"tinc-web-boot/cmd/tinc-web-boot/internal"
	"tinc-web-boot/network"
	"tinc-web-boot/tincd"
	"tinc-web-boot/web"
)

type Main struct {
	Run    Root   `cmd:"run" default:"1"`
	Subnet Subnet `cmd:"subnet"`
}

type Subnet struct {
	Add    AddSubnet    `cmd:"add"`
	Remove RemoveSubnet `cmd:"remove"`
}

type AddSubnet struct {
	Subnet string `name:"subnet" env:"SUBNET" help:"Subnet address" required:"yes"`
	Node   string `name:"node" env:"NODE" help:"PeerInfo name" required:"yes"`
}

type RemoveSubnet struct {
	Node string `name:"node" env:"NODE" help:"PeerInfo name" required:"yes"`
}

type Root struct {
	TincBin string `name:"tinc-bin" env:"TINC_BIN" help:"Custom tinc binary location" default:"tincd"`
	Host    string `name:"host" env:"HOST" help:"Binding host" default:"127.0.0.1"`
	Dir     string `name:"dir" env:"DIR" help:"Directory for config" default:"networks"`
	Dev     bool   `name:"dev" env:"DEV" help:"Enable DEV mode (CORS + logging)"`
	internal.HttpServer
}

func main() {
	var cli Main
	ctx := kong.Parse(&cli)
	err := ctx.Run(nil)
	ctx.FatalIfErrorf(err)
}

func (m *Root) Run() error {
	if !m.Dev {
		gin.SetMode(gin.ReleaseMode)
	}

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

	pool, err := tincd.New(ctx, stor, binary)
	if err != nil {
		return err
	}
	defer pool.Stop()

	_, err = pool.Create("test")
	if err != nil {
		return err
	}

	webApi := web.New(pool, m.Dev)

	return m.Serve(ctx, webApi)
}

func (m *AddSubnet) Run() error {
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

	ntw := &network.Network{Root: "."}
	config, err := ntw.Read()
	if err != nil {
		return err
	}
	selfNode, err := ntw.Node(config.Name)
	if err != nil {
		return err
	}
	address := strings.TrimSpace(strings.Split(selfNode.Subnet, "/")[0])

	url := "http://" + address + ":" + strconv.Itoa(network.CommunicationPort) + "/rpc/watch"
	for {

		err := post(ctx, url, map[string]string{
			"node":   m.Node,
			"subnet": m.Subnet,
		})

		if err != nil {
			log.Println(err)
		} else {
			break
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(time.Second):
		}
	}
	return nil
}

func (m *RemoveSubnet) Run() error {
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
	ntw := &network.Network{Root: "."}
	config, err := ntw.Read()
	if err != nil {
		return err
	}
	selfNode, err := ntw.Node(config.Name)
	if err != nil {
		return err
	}
	address := strings.TrimSpace(strings.Split(selfNode.Subnet, "/")[0])

	url := "http://" + address + ":" + strconv.Itoa(network.CommunicationPort) + "/rpc/forget"
	for {

		err := post(ctx, url, map[string]string{
			"node": m.Node,
		})

		if err != nil {
			log.Println(err)
		} else {
			break
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(time.Second):
		}
	}
	return nil
}

func post(ctx context.Context, URL string, data interface{}) error {
	bdata, err := json.Marshal(data)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, URL, bytes.NewReader(bdata))
	if err != nil {
		return err
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK && res.StatusCode != http.StatusNoContent {
		return fmt.Errorf(res.Status)
	}

	return nil
}
