package network

import (
	"bytes"
	"context"
	"encoding/binary"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"
	"tinc-web-boot/utils"
)

type Storage struct {
	Root string
}

func (st *Storage) Init() error {
	abs, err := filepath.Abs(st.Root)
	if err != nil {
		return err
	}
	st.Root = abs
	err = os.MkdirAll(st.Root, 0755)
	if err != nil {
		return err
	}
	return ApplyOwnerOfSudoUser(st.Root)
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
	lock sync.Mutex
}

func (network *Network) Name() string {
	return filepath.Base(network.Root)
}

func (network *Network) Update(config *Config) error {
	err := os.MkdirAll(network.hosts(), 0755)
	if err != nil {
		return err
	}
	err = ApplyOwnerOfSudoUser(network.hosts())
	if err != nil {
		return err
	}
	err = ApplyOwnerOfSudoUser(network.Root)
	if err != nil {
		return err
	}
	data, err := config.Build()
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(network.configFile(), data, 0755)
	if err != nil {
		return err
	}
	return ApplyOwnerOfSudoUser(network.configFile())
}

func (network *Network) Read() (*Config, error) {
	return ConfigFromFile(network.configFile())
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
	data, err := ioutil.ReadFile(network.NodeFile(name))
	if err != nil {
		return nil, err
	}
	var nd Node
	return &nd, nd.Parse(data)
}

func (network *Network) Self() (*Node, error) {
	cfg, err := network.Read()
	if err != nil {
		return nil, err
	}
	return network.Node(cfg.Name)
}

func (network *Network) Upgrade(upgrade Upgrade) error {
	config, err := network.Read()
	if err != nil {
		return err
	}
	n, err := network.Node(config.Name)
	if err != nil {
		return err
	}
	n.Version = n.Version + 1
	if upgrade.Address != nil {
		n.Address = upgrade.Address
	}
	if upgrade.Port != 0 {
		n.Port = upgrade.Port
		config.Port = upgrade.Port
	}
	if upgrade.Device != "" {
		config.Device = upgrade.Device
	}
	if err := network.Update(config); err != nil {
		return err
	}
	return network.put(n)
}

func (network *Network) Put(node *Node) error {
	if !IsValidNodeName(node.Name) {
		return fmt.Errorf("invalid node name")
	}
	if node.PublicKey == "" {
		return fmt.Errorf("empty public key")
	}
	if node.Subnet == "" {
		return fmt.Errorf("empty subnet")
	}
	network.lock.Lock()
	defer network.lock.Unlock()
	if n, err := network.Node(node.Name); err == nil && n.Version >= node.Version {
		// no need to update - saved version is bigger or equal
		return nil
	}
	config, err := network.Read()
	if err != nil {
		return err
	}
	if config.Name == node.Name {
		// do not touch self node host file
		return nil
	}
	self, err := network.Node(config.Name)
	if err != nil {
		return err
	}
	if self.Subnet != node.Subnet {
		return fmt.Errorf("missmatch subnet for self node (%s) and new node %s (%s)", self.Subnet, node.Name, node.Subnet)
	}
	return network.put(node)
}

func (network *Network) put(node *Node) error {
	data, err := node.Build()
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(network.NodeFile(node.Name), data, 0755)
	if err != nil {
		return err
	}
	return ApplyOwnerOfSudoUser(network.NodeFile(node.Name))
}

func (network *Network) IsDefined() bool {
	v, err := os.Stat(network.configFile())
	return err == nil && !v.IsDir()
}

func (network *Network) Configure(ctx context.Context, tincBin string, subnet *net.IPNet) error {
	if !IsValidName(network.Name()) {
		return fmt.Errorf("invalid network name")
	}
	if err := os.MkdirAll(network.hosts(), 0755); err != nil {
		return err
	}
	if err := network.defineConfiguration(subnet); err != nil {
		return err
	}
	return network.Prepare(ctx, tincBin)
}

func (network *Network) Prepare(ctx context.Context, tincBin string) error {
	config, err := network.Read()
	if err != nil {
		return err
	}

	if err := network.saveScript("tinc-up", tincUp(config)); err != nil {
		return err
	}
	if err := network.saveScript("tinc-down", tincDown(config)); err != nil {
		return err
	}

	if err := network.generateKeysIfNeeded(ctx, tincBin); err != nil {
		return fmt.Errorf("%s: generate keys: %w", network.Name(), err)
	}
	if err := ApplyOwnerOfSudoUser(network.privateKeyFile()); err != nil {
		return fmt.Errorf("apply sudo user on private key: %w", err)
	}
	if err := network.indexPublicNodes(); err != nil {
		return fmt.Errorf("%s: index public nodes: %w", network.Name(), err)
	}
	return network.postConfigure(ctx, config, tincBin)
}

func (network *Network) Logfile() string {
	return filepath.Join(network.Root, "log.txt")
}

func (network *Network) Pidfile() string {
	return filepath.Join(network.Root, "pid.run")
}

func (network *Network) Destroy() error {
	return os.RemoveAll(network.Root)
}

func (network *Network) indexPublicNodes() error {
	config, err := network.Read()
	if err != nil {
		return err
	}

	var publicNodes []string

	list, err := network.Nodes()
	if err != nil {
		return err
	}

	for _, node := range list {
		info, err := network.Node(node)
		if err != nil {
			return fmt.Errorf("parse node %s: %w", node, err)
		}
		if len(info.Address) > 0 {
			publicNodes = append(publicNodes, node)
		}
	}

	config.ConnectTo = publicNodes

	return network.Update(config)
}

func (network *Network) defineConfiguration(subnet *net.IPNet) error {
	if network.IsDefined() {
		return nil
	}
	mask, bits := subnet.Mask.Size()
	if bits != 32 {
		return fmt.Errorf("currently supported only IPv4 subnets (%d bits) but requested %d bits", 32, bits)
	}
	hostname, _ := os.Hostname()
	suffix := utils.RandStringRunesCustom(6, suffixRunes)
	nodeName := regexp.MustCompile(`[^a-z0-9]*`).ReplaceAllString(strings.ToLower(hostname), "") + "_" + suffix
	selfIP := generateRandomIPv4(subnet)
	config := &Config{
		Name:      nodeName,
		Port:      uint16(30000 + rand.Intn(35535)),
		Interface: "tinc" + suffix,
		AutoStart: false,
		Mode:      "switch",
		IP:        selfIP.String(),
		Mask:      mask,
		Broadcast: "mst",
	}

	if err := network.beforeConfigure(config); err != nil {
		return err
	}

	if err := network.Update(config); err != nil {
		return err
	}

	var version = 1
	if n, err := network.Node(config.Name); err == nil {
		version = n.Version + 1
	}

	nodeConfig := &Node{
		Name:    nodeName,
		Subnet:  subnet.String(),
		Port:    config.Port,
		Version: version,
	}

	return network.put(nodeConfig)
}

func (network *Network) configFile() string {
	return filepath.Join(network.Root, "tinc.conf")
}

func (network *Network) hosts() string {
	return filepath.Join(network.Root, "hosts")
}

func (network *Network) NodeFile(name string) string {
	name = regexp.MustCompile(`^[^a-zA-Z0-9_]+$`).ReplaceAllString(name, "")
	return filepath.Join(network.hosts(), name)
}

func (network *Network) scriptFile(name string) string {
	return filepath.Join(network.Root, name+scriptSuffix)
}

func (network *Network) privateKeyFile() string {
	return filepath.Join(network.Root, "rsa_key.priv")
}

func (network *Network) saveScript(name string, content string) error {
	file := network.scriptFile(name)
	err := ioutil.WriteFile(file, []byte(content), 0755)
	if err != nil {
		return fmt.Errorf("%s: generate script %s: %w", network.Name(), name, err)
	}
	err = postProcessScript(file)
	if err != nil {
		return fmt.Errorf("%s: post-process script %s: %w", network.Name(), name, err)
	}
	return nil
}

func (network *Network) generateKeysIfNeeded(ctx context.Context, tincBin string) error {
	_, err := os.Stat(network.privateKeyFile())
	if err == nil {
		return nil
	}
	if !os.IsNotExist(err) {
		return err
	}

	cmd := exec.CommandContext(ctx, tincBin, "-K", "4096", "-c", network.Root)
	cmd.Stdin = bytes.NewReader(nil)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	utils.SetCmdAttrs(cmd)

	return cmd.Run()
}

var suffixRunes = []rune("ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")

func IsValidName(name string) bool {
	return regexp.MustCompile(`^[a-zA-Z0-9_-]+$`).MatchString(name)
}

func IsValidNodeName(name string) bool {
	return regexp.MustCompile(`^[a-zA-Z0-9_]+$`).MatchString(name)
}

func generateRandomIPv4(subnet *net.IPNet) net.IP {
	netsize, bits := subnet.Mask.Size()
	baseIP := binary.BigEndian.Uint32(subnet.IP)

	maxIP := (1 << (bits - netsize)) - 1
	if maxIP <= 0 {
		return subnet.IP
	}
	genIP := baseIP + (1 + rand.Uint32()%uint32(maxIP))

	var ip [4]byte
	binary.BigEndian.PutUint32(ip[:], uint32(genIP))
	return net.IP(ip[:])
}

func init() {
	rand.Seed(time.Now().UnixNano())
}
