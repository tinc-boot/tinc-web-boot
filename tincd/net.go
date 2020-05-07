package tincd

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"
	"tinc-web-boot/network"
	"tinc-web-boot/tincd/beacon"
	"tinc-web-boot/utils"
)

const (
	beaconPort = 2655
	beaconText = "tinc-web-boot i-am-here"
)

type netImpl struct {
	tincBin string
	ctx     context.Context

	done   chan struct{}
	stop   func()
	lock   sync.Mutex
	peers  peersManager
	events *network.Events

	definition *network.Network
}

func (impl *netImpl) Start() {
	impl.lock.Lock()
	defer impl.lock.Unlock()
	impl.unsafeStop()

	ctx, cancel := context.WithCancel(impl.ctx)
	done := make(chan struct{})
	impl.stop = cancel
	impl.done = done
	impl.peers = peersManager{
		network: impl.definition,
		events:  impl.events,
	}
	go func() {
		defer cancel()
		defer close(done)
		defer impl.events.Stopped.Emit(network.NetworkID{Name: impl.definition.Name()})
		err := impl.run(ctx)
		if err != nil {
			log.Println("failed run network", impl.definition.Name(), ":", err)
		}
	}()
}

func (impl *netImpl) Stop() {
	impl.lock.Lock()
	defer impl.lock.Unlock()
	impl.unsafeStop()
}

func (impl *netImpl) Peers() []*Peer {
	return impl.peers.List()
}

func (impl *netImpl) Peer(name string) (*Peer, bool) {
	return impl.peers.Get(name)
}

func (impl *netImpl) Definition() *network.Network {
	return impl.definition
}

func (impl *netImpl) IsRunning() bool {
	ch := impl.done
	if ch == nil {
		return false
	}
	select {
	case <-ch:
		return false
	default:
		return true
	}
}

func (impl *netImpl) unsafeStop() {
	v := impl.stop
	if v == nil {
		return
	}
	v()
	<-impl.done
	impl.stop = nil
}

func (impl *netImpl) run(global context.Context) error {
	if err := impl.definition.Prepare(global, impl.tincBin); err != nil {
		return fmt.Errorf("configure: %w", err)
	}

	absDir, err := filepath.Abs(impl.definition.Root)
	if err != nil {
		return err
	}

	config, err := impl.definition.Read()
	if err != nil {
		return err
	}
	interfaceName := config.Interface
	if interfaceName == "" { // for darwin
		interfaceName = config.Device[strings.LastIndex(config.Device, "/")+1:]
	}

	ctx, abort := context.WithCancel(global)
	defer abort()

	cmd := exec.CommandContext(ctx, impl.tincBin, "-D", "-d", "-d", "-d",
		"--pidfile", impl.definition.Pidfile(),
		"--logfile", impl.definition.Logfile(),
		"-c", absDir)
	cmd.Dir = absDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	utils.SetCmdAttrs(cmd)

	peers := make(chan peerReq)

	var wg sync.WaitGroup

	// run tinc service
	wg.Add(1)
	go func() {
		defer wg.Done()
		defer abort()
		err := cmd.Run()
		if err != nil {
			log.Println(impl.definition.Name(), "failed to run tinc:", err)
		}
	}()

	// run http API
	wg.Add(1)
	go func() {
		defer wg.Done()
		defer abort()
		runAPI(ctx, impl.definition)
		log.Println(impl.definition.Name(), "api stopped")
	}()

	// run broadcaster (to find another nodes)
	wg.Add(1)
	go func() {
		defer wg.Done()
		defer abort()
		for {
			err := impl.runBroadcaster(ctx, interfaceName, peers)
			if err != nil {
				log.Println("failed start broadcaster:", err, "interface:", interfaceName)
			}
			select {
			case <-ctx.Done():
				return
			case <-time.After(1 * time.Second):
			}
		}
	}()

	// run peers checker
	wg.Add(1)
	go func() {
		defer wg.Done()
		defer abort()
		impl.peers.Run(ctx, peers)
	}()

	// run periodic query of peers
	wg.Add(1)
	go func() {
		defer wg.Done()
		defer abort()
		impl.queryActivePeers(ctx)
	}()

	// fix: change owner of log file and pid file to process runner
	wg.Add(1)
	go func() {
		defer wg.Done()
		select {
		case <-ctx.Done():
		case <-time.After(2 * time.Second):
			_ = network.ApplyOwnerOfSudoUser(impl.definition.Logfile())
			_ = network.ApplyOwnerOfSudoUser(impl.definition.Pidfile())
		}
	}()

	impl.events.Started.Emit(network.NetworkID{Name: impl.definition.Name()})
	wg.Wait()
	return ctx.Err()
}

func (impl *netImpl) runBroadcaster(ctx context.Context, interfaceName string, peers chan<- peerReq) error {
	beacons, err := beacon.Run(ctx, interfaceName, beaconText, beaconPort)
	if err != nil {

		return err
	}
	log.Println("[TRACE]", "broadcaster started on", interfaceName)
	filtered := beacon.FilterByContent(ctx, beacons, []byte(beaconText))
LOOP:
	for update := range beacon.Discovery(ctx, filtered, beacon.DefaultKeepAlive*2) {
		log.Println("[TRACE]", "found beacon from", update.Address, "action:", update.Action)
		if update.Action == beacon.Updated {
			continue
		}
		addr, _, _ := net.SplitHostPort(update.Address)
		select {
		case peers <- peerReq{
			Address: addr,
			Add:     update.Action == beacon.Discovered,
		}:
		case <-ctx.Done():
			break LOOP
		}
	}
	return nil
}

func (impl *netImpl) queryActivePeers(ctx context.Context) {
	for {
		for _, peer := range impl.peers.List() {
			list, err := peer.fetchNodes(ctx)
			if err != nil {
				log.Println("failed to fetch list of nodes from", peer.Node, ":", err)
				continue
			}

			for _, node := range list {
				err = impl.Definition().Put(node)
				if err != nil {
					log.Println("failed to save node", node.Name, ":", err)
					continue
				}
			}
		}

		select {
		case <-time.After(nodesListInterval):
		case <-ctx.Done():
			return
		}

	}
}
