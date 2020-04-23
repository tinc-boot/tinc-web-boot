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
	Name string `arg:"name" required:"yes"`
}

func (m *shareNetwork) Run(global *globalContext) error {
	share, err := m.Client().Share(global.ctx, m.Name)
	if err != nil {
		return err
	}
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(share)
}
