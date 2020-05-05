package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/alecthomas/kong"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"time"
	"tinc-web-boot/cmd/tinc-web-boot/internal"
	"tinc-web-boot/network"
	"tinc-web-boot/tincd"
	"tinc-web-boot/web"
)

var version = "dev"

const (
	configFile = "tinc-web.json"
)

type globalContext struct {
	ctx context.Context
}

type Main struct {
	Run     Root             `cmd:"run" default:"1" json:"run"`
	New     create           `cmd:"new" help:"Create new network"  json:"-"`
	Delete  remove           `cmd:"delete" help:"Delete network"  json:"-"`
	Join    join             `cmd:"join" help:"Join by majordomo"  json:"-"`
	Invite  invite           `cmd:"invite" help:"Invite people by link"  json:"-"`
	List    listNetworks     `cmd:"list" help:"List networks"  json:"-"`
	Info    getNetwork       `cmd:"info" help:"Get network info"  json:"-"`
	Share   shareNetwork     `cmd:"share" help:"Share network"  json:"-"`
	Import  importNetwork    `cmd:"import" help:"Import network"  json:"-"`
	Start   start            `cmd:"start" help:"Start network"  json:"-"`
	Stop    stop             `cmd:"stop" help:"Stop network"  json:"-"`
	Peers   peers            `cmd:"peers" help:"List connected peers"  json:"-"`
	Upgrade upgrade          `cmd:"upgrade" help:"Upgrade network"  json:"-"`
	Version kong.VersionFlag `name:"version" help:"print version and exit"  json:"-"`
}

type Root struct {
	TincBin         string   `name:"tinc-bin" env:"TINC_BIN" help:"Custom tinc binary location" default:"tincd" json:"tinc_bin"`
	Host            string   `name:"host" env:"HOST" help:"Binding host" default:"127.0.0.1" json:"host"`
	Dir             string   `name:"dir" env:"DIR" help:"Directory for config" default:"networks" json:"dir"`
	Dev             bool     `name:"dev" env:"DEV" help:"Enable DEV mode (CORS + logging)" json:"-"`
	Headless        bool     `long:"headless" env:"HEADLESS" description:"Disable launch browser" json:"-"`
	DevGenOnly      bool     `name:"dev-gen-only" env:"DEV_GEN_ONLY" help:"(dev only) generate sample config but don't run" json:"-"`
	DevNet          string   `name:"dev-net" env:"DEV_NET" help:"(dev only) Name of development network" default:"example-network" json:"-"`
	DevAddress      []string `name:"dev-address" env:"DEV_ADDRESS" help:"(dev only) Public addresses" default:"127.0.0.1" json:"-"`
	DevPort         uint16   `name:"dev-port" env:"DEV_PORT" help:"(dev only) Development port" default:"10655" json:"-"`
	DevSubnet       string   `name:"dev-subnet" env:"DEV_SUBNET" help:"(dev only) Custom subnet for sample network" default:"10.155.0.0/16" json:"-"`
	NoApp           bool     `name:"no-app" env:"NO_APP" help:"Don't try to open UI in application mode (if possible)" json:"no_app"`
	UIPublicAddress []string `short:"A" name:"ui-public-address" env:"UI_PUBLIC_ADDRESS" help:"Custom UI public addresses (host:port) for links" json:"ui_public_addresses"`
	AuthKey         string   `name:"auth-key" env:"AUTH_KEY" help:"JWT signing key (empty - autogenerated)" json:"auth_key"`
	DumpKey         string   `short:"f" name:"dump-key" env:"DUMP_KEY" help:"Dump API token" default:".tinc-web-boot" json:"dump_key"`
	internal.HttpServer
}

func main() {
	var cli Main
	ctx := kong.Parse(&cli, kong.Vars{"version": version}, kong.Configuration(kong.JSON, configFile))
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		if err := cli.dumpConfig(configFile); err != nil {
			log.Println("[WARN]", "failed dump config file", configFile, ":", err)
		}
	}
	gctx, closer := context.WithCancel(context.Background())
	go func() {
		c := make(chan os.Signal, 2)
		signal.Notify(c, os.Kill, os.Interrupt)
		for range c {
			closer()
			break
		}
	}()
	defer closer()
	err := ctx.Run(&globalContext{ctx: gctx})
	ctx.FatalIfErrorf(err)
}

func (cli Main) dumpConfig(filename string) error {
	data, err := json.MarshalIndent(cli, "", "  ")
	if err != nil {
		return err
	}
	return ioutil.WriteFile(filename, data, 0755)
}

func (m *Root) Run(global *globalContext) error {
	if !m.Dev {
		gin.SetMode(gin.ReleaseMode)
	}

	binary, err := internal.DetectTincBinary(m.TincBin)
	if err != nil {
		return err
	}
	log.Println("detected Tinc binary:", binary)

	err = internal.Preload(global.ctx)
	if err != nil {
		return err
	}
	log.Println("preload complete")

	stor := &network.Storage{Root: m.Dir}
	err = stor.Init()
	if err != nil {
		return err
	}

	pool, err := tincd.New(global.ctx, stor, binary)
	if err != nil {
		return err
	}
	defer pool.Stop()

	pool.Events().Sink(func(eventName string, payload interface{}) {
		log.Printf("[TRACE] (%s) %+v", eventName, payload)
	})

	if m.Dev {
		_, subnet, err := net.ParseCIDR(m.DevSubnet)
		if err != nil {
			return err
		}
		ntw, err := pool.Create(m.DevNet, subnet)
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
	if m.AuthKey == "" {
		m.AuthKey = uuid.New().String()
	}
	_, portStr, _ := net.SplitHostPort(m.Bind)
	port, _ := strconv.Atoi(portStr)
	apiCfg := web.Config{
		Dev:             m.Dev,
		AuthorizedOnly:  m.Headless,
		AuthKey:         m.AuthKey,
		LocalUIPort:     uint16(port),
		PublicAddresses: m.UIPublicAddress,
	}
	webApi, uiApp := apiCfg.New(pool)
	if !m.Headless {
		go func() {

			for i := 0; i < 50; i++ {
				if isGuiAvailable(global.ctx, m.Bind, time.Second) {
					break
				}
				select {
				case <-time.After(100 * time.Millisecond):
				case <-global.ctx.Done():
					return
				}
			}

			err := internal.OpenInBrowser(global.ctx, "http://"+m.Bind, !m.NoApp)
			if err != nil {
				log.Println("failed to open UI:", err)
			} else {
				log.Println("UI opened")
			}
		}()
	} else {
		token, err := uiApp.IssueAccessToken(3650)
		if err != nil {
			return fmt.Errorf("issue token: %w", err)
		}
		fmt.Println("\n-------------\n\n", "TOKEN:", token, "\n\n-------------")
		if m.DumpKey != "" {
			err = ioutil.WriteFile(m.DumpKey, []byte(token), 0755)
			if err != nil {
				log.Println("[WARN]", "failed to dump key:", err)
			}
		}
	}
	return m.Serve(global.ctx, webApi)
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
