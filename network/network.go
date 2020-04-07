package network

import (
	"context"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
)

type Storage struct {
	Root string
}

func (st *Storage) Init() error {
	return os.MkdirAll(st.Root, 0755)
}

func (st *Storage) Get(network string) *Network {
	return &Network{Root: st.WorkDir(network)}
}

func (st *Storage) List() ([]*Network, error) {
	ls, err := ioutil.ReadDir(st.Root)
	if err != nil {
		return nil, err
	}
	var ans = make([]*Network, 0, len(ls))
	for _, l := range ls {
		if l.IsDir() {
			ans = append(ans, &Network{Root: filepath.Join(st.Root, l.Name())})
		}
	}
	return ans, nil
}

func (st *Storage) WorkDir(network string) string {
	network = regexp.MustCompile(`^[^a-zA-Z0-9_]+$`).ReplaceAllString(network, "")
	return filepath.Join(st.Root, network)
}

type Network struct {
	Root string
}

func (network *Network) Name() string {
	return filepath.Base(network.Root)
}

func (network *Network) Update(config *Config) error {
	err := os.MkdirAll(network.hosts(), 0755)
	if err != nil {
		return err
	}
	data, err := config.MarshalText()
	if err != nil {
		return err
	}
	return ioutil.WriteFile(network.configFile(), data, 0755)
}

func (network *Network) Read() (*Config, error) {
	data, err := ioutil.ReadFile(network.configFile())
	if err != nil {
		return nil, err
	}
	var cfg Config
	return &cfg, cfg.UnmarshalText(data)
}

func (network *Network) Nodes() ([]string, error) {
	list, err := ioutil.ReadDir(network.hosts())
	if err != nil {
		return nil, err
	}
	var ans = make([]string, 0, len(list))
	for _, v := range list {
		if !v.IsDir() {
			ans = append(ans, v.Name())
		}
	}
	return ans, nil
}

func (network *Network) Node(name string) (*Node, error) {
	data, err := ioutil.ReadFile(network.node(name))
	if err != nil {
		return nil, err
	}
	var nd Node
	return &nd, nd.UnmarshalText(data)
}

func (network *Network) configure(ctx context.Context, apiPort int, tincBin string) error {
	config, err := network.Read()
	if err != nil {
		return err
	}
	selfNode, err := network.Node(config.Name)
	if err != nil {
		return err
	}
	selfExec, err := os.Executable()
	if err != nil {
		return err
	}
	if err := network.saveScript("tinc-up", tincUp(selfNode)); err != nil {
		return err
	}
	if err := network.saveScript("tinc-down", tincDown(selfNode)); err != nil {
		return err
	}
	if err := network.saveScript("subnet-up", subnetUp(selfExec, apiPort)); err != nil {
		return err
	}
	if err := network.saveScript("subnet-down", subnetDown(selfExec, apiPort)); err != nil {
		return err
	}
	return network.postConfigure(ctx, config, tincBin)
}

func (network *Network) configFile() string {
	return filepath.Join(network.Root, "tinc.conf")
}

func (network *Network) hosts() string {
	return filepath.Join(network.Root, "hosts")
}

func (network *Network) node(name string) string {
	name = regexp.MustCompile(`^[^a-zA-Z0-9_]+$`).ReplaceAllString(name, "")
	return filepath.Join(network.hosts(), name)
}

func (network *Network) scriptFile(name string) string {
	return filepath.Join(network.Root, name+scriptSuffix)
}

func (network *Network) saveScript(name string, content string) error {
	file := network.scriptFile(name)
	err := ioutil.WriteFile(name, []byte(content), 0755)
	if err != nil {
		return err
	}
	return postProcessScript(file)
}
