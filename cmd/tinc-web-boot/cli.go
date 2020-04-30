package main

import (
	"encoding/json"
	"fmt"
	"github.com/olekukonko/tablewriter"
	"os"
	"tinc-web-boot/support/go/tincweb"
)

type baseParam struct {
	URL   string `name:"url" env:"URL" help:"API URL for tinc-web-boot" default:"http://127.0.0.1:8686/api"`
	Token string `name:"token" env:"TOKEN" help:"Access token for API" default:"local"`
}

func (bp baseParam) Client() *tincweb.TincWebClient {
	return &tincweb.TincWebClient{BaseURL: bp.URL + "/" + bp.Token}
}

type listNetworks struct {
	baseParam
}

func (m *listNetworks) Run(global *globalContext) error {
	list, err := m.Client().Networks(global.ctx)
	if err != nil {
		return err
	}
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Network", "Running"})
	for _, net := range list {
		table.Append([]string{
			net.Name, fmt.Sprint(net.Running),
		})
	}
	table.Render()
	return nil
}

type getNetwork struct {
	baseParam
	Name string `arg:"name" required:"yes"`
}

func (m *getNetwork) Run(global *globalContext) error {
	info, err := m.Client().Network(global.ctx, m.Name)
	if err != nil {
		return err
	}
	fmt.Println("Name:", info.Name)
	fmt.Println("Running:", info.Running)
	if info.Config == nil {
		return nil
	}
	fmt.Println("IP:", info.Config.IP)
	fmt.Println("Mask:", info.Config.Mask)
	fmt.Println("Node:", info.Config.Name)
	fmt.Println("Device:", info.Config.Device)
	fmt.Println("Device type:", info.Config.DeviceType)
	fmt.Println("Interface:", info.Config.Interface)
	fmt.Println("Port:", info.Config.Port)
	fmt.Println("Mode:", info.Config.Mode)
	fmt.Println("Autostart:", info.Config.AutoStart)
	for _, c := range info.Config.ConnectTo {
		fmt.Println("Connect to:", c)
	}
	return nil
}

type shareNetwork struct {
	baseParam
	Output string `short:"o" name:"output" env:"OUTPUT" help:"Output file (empty or - for stdout)" default:"-"`
	Name   string `arg:"name" required:"yes"`
}

func (m *shareNetwork) Run(global *globalContext) error {
	share, err := m.Client().Share(global.ctx, m.Name)
	if err != nil {
		return err
	}
	var f = os.Stdout
	if m.Output != "" && m.Output != "-" {
		fs, err := os.Create(m.Output)
		if err != nil {
			return err
		}
		defer fs.Close()
		f = fs
	}
	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	return enc.Encode(share)
}

type importNetwork struct {
	baseParam
	Input string `short:"i" name:"input" env:"INPUT" help:"Input file (empty or - for stdin)" default:"-"`
	Name  string `arg:"name" help:"optional name for network" optional:"yes"`
}

func (m *importNetwork) Run(global *globalContext) error {
	var f = os.Stdin
	if m.Input != "" && m.Input != "-" {
		fs, err := os.Open(m.Input)
		if err != nil {
			return err
		}
		defer fs.Close()
		f = fs
	}
	dec := json.NewDecoder(f)
	var cfg tincweb.Sharing
	err := dec.Decode(&cfg)
	if err != nil {
		return err
	}
	if m.Name != "" {
		cfg.Name = m.Name
	}
	_, err = m.Client().Import(global.ctx, cfg)
	return err
}

type peers struct {
	baseParam
	Network string `arg:"network" required:"yes"`
}

func (m *peers) Run(global *globalContext) error {
	list, err := m.Client().Peers(global.ctx, m.Network)
	if err != nil {
		return err
	}
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Name", "Connected", "Address", "Version"})
	for _, peer := range list {
		info, err := m.Client().Peer(global.ctx, m.Network, peer.Name)
		if err != nil {
			return err
		}
		var addr string
		if peer.Status != nil {
			addr = peer.Status.Address
		}

		table.Append([]string{
			peer.Name, fmt.Sprint(peer.Online), addr, fmt.Sprint(info.Configuration.Version),
		})
	}
	table.Render()
	return nil
}

type start struct {
	baseParam
	Network string `arg:"network" required:"yes"`
}

func (m *start) Run(global *globalContext) error {
	ntw, err := m.Client().Start(global.ctx, m.Network)
	if err != nil {
		return err
	}
	fmt.Println("name:", ntw.Name, "running:", ntw.Running)
	return nil
}

type stop struct {
	baseParam
	Network string `arg:"network" required:"yes"`
}

func (m *stop) Run(global *globalContext) error {
	ntw, err := m.Client().Stop(global.ctx, m.Network)
	if err != nil {
		return err
	}
	fmt.Println("name:", ntw.Name, "running:", ntw.Running)
	return nil
}
