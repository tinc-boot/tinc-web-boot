package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
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

var version = "dev"

type Main struct {
	Run     Root             `cmd:"run" default:"1"`
	Subnet  Subnet           `cmd:"subnet"`
	Version kong.VersionFlag `name:"version" help:"print version and exit"`
}

type Subnet struct {
	Add    AddSubnet    `cmd:"add"`
	Remove RemoveSubnet `cmd:"remove"`
}

type AddSubnet struct {
	Subnet  string `name:"subnet" env:"SUBNET" help:"Subnet address" required:"yes"`
	Node    string `name:"node" env:"NODE" help:"PeerInfo name" required:"yes"`
	Retries int    `name:"retries" env:"RETRIES" help:"Retries attempts" default:"5"`
}

type RemoveSubnet struct {
	Node    string `name:"node" env:"NODE" help:"PeerInfo name" required:"yes"`
	Retries int    `name:"retries" env:"RETRIES" help:"Retries attempts" default:"5"`
}

type Root struct {
	TincBin    string   `name:"tinc-bin" env:"TINC_BIN" help:"Custom tinc binary location" default:"tincd"`
	Host       string   `name:"host" env:"HOST" help:"Binding host" default:"127.0.0.1"`
	Dir        string   `name:"dir" env:"DIR" help:"Directory for config" default:"networks"`
	Dev        bool     `name:"dev" env:"DEV" help:"Enable DEV mode (CORS + logging)"`
	Headless   bool     `long:"headless" env:"HEADLESS" description:"Disable launch browser"`
	DevGenOnly bool     `name:"dev-gen-only" env:"DEV_GEN_ONLY" help:"(dev only) generate sample config but don't run"`
	DevNet     string   `name:"dev-net" env:"DEV_NET" help:"(dev only) Name of development network" default:"example-network"`
	DevAddress []string `name:"dev-address" env:"DEV_ADDRESS" help:"(dev only) Public addresses" default:"127.0.0.1"`
	DevPort    uint16   `name:"dev-port" env:"DEV_PORT" help:"(dev only) Development port" default:"10655"`
	DevSubnet  string   `name:"dev-subnet" env:"DEV_SUBNET" help:"(dev only) Custom subnet for sample network (empty is random)"`
	NoApp      bool     `name:"no-app" env:"NO_APP" help:"Don't try to open UI in application mode (if possible)"`
	internal.HttpServer
}

func main() {
	var cli Main
	ctx := kong.Parse(&cli, kong.Vars{"version": version})
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
	log.Println("detected Tinc binary:", binary)

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

	err = internal.Preload(ctx)
	if err != nil {
		return err
	}
	log.Println("preload complete")

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

	if m.Dev {
		ntw, err := pool.Create(m.DevNet)
		if err != nil {
			return err
		}
		var addrs []network.Address
		for _, addr := range m.DevAddress {
			addrs = append(addrs, network.Address{
				Host: addr,
				Port: m.DevPort,
			})
		}
		err = ntw.Definition().Upgrade(network.Upgrade{
			Address: addrs,
			Port:    m.DevPort,
			Subnet:  m.DevSubnet,
		})
		if err != nil {
			return err
		}
		if m.DevGenOnly {
			return nil
		}
		if !ntw.IsRunning() {
			ntw.Start()
		}
	}

	webApi := web.New(pool, m.Dev, m.Dev || !m.Headless)
	if !m.Headless {
		go func() {

			for i := 0; i < 50; i++ {
				if isGuiAvailable(ctx, m.Bind, time.Second) {
					break
				}
				select {
				case <-time.After(100 * time.Millisecond):
				case <-ctx.Done():
					return
				}
			}

			err := internal.OpenInBrowser(ctx, "http://"+m.Bind, !m.NoApp)
			if err != nil {
				log.Println("failed to open UI:", err)
			} else {
				log.Println("UI opened")
			}
		}()
	}
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
	for i := 0; i < m.Retries; i++ {

		err := post(ctx, url, map[string]string{
			"node":   m.Node,
			"subnet": m.Subnet,
		})
		if err == badRequest {
			return badRequest
		}
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
	for i := 0; i < m.Retries; i++ {

		err := post(ctx, url, map[string]string{
			"node": m.Node,
		})

		if err == badRequest {
			return badRequest
		}

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

var badRequest = errors.New("bad request")

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
	if res.StatusCode == http.StatusBadRequest {
		return badRequest
	}
	if res.StatusCode != http.StatusOK && res.StatusCode != http.StatusNoContent {
		return fmt.Errorf(res.Status)
	}

	return nil
}

func isGuiAvailable(global context.Context, url string, timeout time.Duration) bool {
	ctx, cancel := context.WithTimeout(global, timeout)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return false
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return false
	}
	res.Body.Close()
	return true
}
