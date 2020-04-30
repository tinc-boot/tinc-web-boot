package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/olekukonko/tablewriter"
	"io/ioutil"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
	"time"
	"tinc-web-boot/support/go/tincweb"
	"tinc-web-boot/support/go/tincwebmajordomo"
)

type baseParam struct {
	URL       string `name:"url" env:"URL" help:"API URL for tinc-web-boot" default:"http://127.0.0.1:8686/api"`
	Token     string `name:"token" env:"TOKEN" help:"Access token for API" default:"local"`
	TokenFile string `short:"f" long:"token-file" env:"TOKEN_FILE" description:"Token file" default:".tinc-web-boot"`
}

func (bp baseParam) Client() *tincweb.TincWebClient {
	if bp.TokenFile != "" && (bp.Token == "" || bp.Token == "local") {
		data, err := ioutil.ReadFile(bp.TokenFile)
		if err == nil {
			bp.Token = string(data)
		}
	}
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
	printNetwork(info)
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

type create struct {
	baseParam
	Network string `arg:"network" required:"yes"`
	Subnet  string `arg:"subnet" required:"yes"`
}

func (m *create) Run(global *globalContext) error {
	info, err := m.Client().Create(global.ctx, m.Network, m.Subnet)
	if err != nil {
		return err
	}
	printNetwork(info)
	return nil
}

type remove struct {
	baseParam
	Network string `arg:"network" required:"yes"`
}

func (m *remove) Run(global *globalContext) error {
	ok, err := m.Client().Remove(global.ctx, m.Network)
	if err != nil {
		return err
	}
	if ok {
		fmt.Println("removed")
	}
	return nil
}

type upgrade struct {
	baseParam
	PublicAddress []string `short:"A" name:"public-address" env:"PUBLIC_ADDRESS" help:"Public node address"`
	Network       string   `arg:"network" required:"yes"`
}

func (m *upgrade) Run(global *globalContext) error {
	var params tincweb.Upgrade
	for _, addr := range m.PublicAddress {
		host, port, err := net.SplitHostPort(addr)
		if err != nil {
			return err
		}
		portV, err := strconv.ParseUint(port, 10, 16)
		if err != nil {
			return err
		}
		params.Address = append(params.Address, tincweb.Address{
			Host: host,
			Port: uint16(portV),
		})
	}

	_, err := m.Client().Upgrade(global.ctx, m.Network, params)
	if err != nil {
		return err
	}
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

type invite struct {
	baseParam
	Lifetime time.Duration `name:"lifetime" env:"LIFETIME" help:"How long invitation will work" default:"1h"`
	Network  string        `arg:"network" required:"yes"`
}

func (m *invite) Run(global *globalContext) error {
	link, err := m.Client().Majordomo(global.ctx, m.Network, m.Lifetime)
	if err != nil {
		return err
	}
	fmt.Println(link)
	return nil
}

type join struct {
	baseParam
	NoStart bool   `name:"no-start" env:"NO_START" help:"Do not start network automatically"`
	URL     string `arg:"url" required:"yes"`
}

func (m *join) Run(global *globalContext) error {
	parts := strings.Split(m.URL, "/")
	token := parts[len(parts)-1]
	data := strings.Split(token, ".")[1]
	bindata, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		return err
	}

	var share struct {
		Network string `json:"network"`
		Subnet  string `json:"subnet"`
	}

	err = json.Unmarshal(bindata, &share)
	if err != nil {
		return err
	}

	remote := &tincwebmajordomo.TincWebMajordomoClient{BaseURL: m.URL}

	ntw, err := m.Client().Create(global.ctx, share.Network, share.Subnet)
	if err != nil {
		return err
	}

	self, err := m.Client().Node(global.ctx, ntw.Name)
	if err != nil {
		return err
	}

	var mapped = &tincwebmajordomo.Node{
		Name:      self.Name,
		Subnet:    self.Subnet,
		Port:      self.Port,
		PublicKey: self.PublicKey,
		Version:   self.Version,
	}
	for _, addr := range self.Address {
		mapped.Address = append(mapped.Address, tincwebmajordomo.Address{
			Host: addr.Host,
			Port: addr.Port,
		})
	}
	shared, err := remote.Join(global.ctx, share.Network, mapped)
	if err != nil {
		return err
	}
	var mappedShared tincweb.Sharing
	err = quickRemap(&mappedShared, shared)
	if err != nil {
		return err
	}

	info, err := m.Client().Import(global.ctx, mappedShared)
	if err != nil {
		return err
	}
	log.Println("SUCCESS!")
	printNetwork(info)
	if !m.NoStart {
		log.Println("Starting...")
		_, err = m.Client().Start(global.ctx, info.Name)
		return err
	}
	return nil
}

func printNetwork(info *tincweb.Network) {
	fmt.Println("Name:", info.Name)
	fmt.Println("Running:", info.Running)
	if info.Config == nil {
		return
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
}

func quickRemap(dst interface{}, src interface{}) error {
	data, err := json.Marshal(src)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, dst)
}
