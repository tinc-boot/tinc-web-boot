package tincd

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"sort"
	"strconv"
	"sync"
	"time"
	"tinc-web-boot/network"
)

type peerReq struct {
	Address string
	Add     bool
}

type peersManager struct {
	lock    sync.RWMutex
	events  *network.Events
	network *network.Network
	list    map[string]*Peer
}

func (pl *peersManager) Run(ctx context.Context, peers <-chan peerReq) {
	var wg sync.WaitGroup
LOOP:
	for {
		select {
		case <-ctx.Done():
			break LOOP
		case req := <-peers:
			if req.Add {
				peer := pl.Add(ctx, req.Address)
				wg.Add(1)
				go func() {
					defer wg.Done()
					pl.events.PeerDiscovered.Emit(network.PeerID{
						Network: pl.network.Name(),
						Node:    peer.Node(),
						Address: peer.Address,
					})
					peer.run(ctx)
					pl.events.PeerJoined.Emit(network.PeerID{
						Network: pl.network.Name(),
						Node:    peer.Node(),
						Address: peer.Address,
					})
				}()
			} else {
				peer := pl.Remove(req.Address)
				if peer != nil {
					pl.events.PeerLeft.Emit(network.PeerID{
						Network: pl.network.Name(),
						Address: peer.Address,
						Node:    peer.Node(),
					})
				}
			}
		}
	}
	wg.Wait()
}

func (pl *peersManager) Add(ctx context.Context, address string) *Peer {
	ctx, cancel := context.WithCancel(ctx)
	p := &Peer{
		Address: address,
		stop:    cancel,
		network: pl.network,
	}
	pl.add(p)
	return p
}

func (pl *peersManager) Remove(address string) *Peer {
	v, ok := pl.remove(address)
	if ok {
		v.stop()
		return v
	}
	return nil
}

func (pl *peersManager) add(newPeer *Peer) {
	pl.lock.Lock()
	defer pl.lock.Unlock()
	if pl.list == nil {
		pl.list = make(map[string]*Peer)
	}
	pl.list[newPeer.Address] = newPeer
}

func (pl *peersManager) remove(address string) (*Peer, bool) {
	pl.lock.Lock()
	defer pl.lock.Unlock()
	v, ok := pl.list[address]
	delete(pl.list, address)
	return v, ok
}

func (pl *peersManager) List() []*Peer {
	pl.lock.RLock()
	var cp = make([]*Peer, 0, len(pl.list))
	for _, v := range pl.list {
		cp = append(cp, v)
	}
	pl.lock.RUnlock()
	sort.Slice(cp, func(i, j int) bool {
		return cp[i].Address < cp[j].Address
	})
	return cp
}

func (pl *peersManager) Get(name string) (*Peer, bool) {
	pl.lock.RLock()
	defer pl.lock.RUnlock()
	v, ok := pl.list[name]
	return v, ok
}

type Peer struct {
	Address string        `json:"address"`
	Fetched bool          `json:"fetched"`
	Config  *network.Node `json:"config,omitempty"`
	stop    func()
	network *network.Network
}

func (peer *Peer) Node() string {
	if peer.Fetched {
		return peer.Config.Name
	}
	return ""
}

func (peer *Peer) run(ctx context.Context) {
	for {
		node, err := peer.fetchConfig(ctx)
		if err == nil {
			err = peer.network.Put(node)
		}
		if err == nil {
			peer.Config = node
			break
		}
		log.Println("failed get", peer.Address, ":", err)
		select {
		case <-ctx.Done():
			return
		case <-time.After(retryInterval):
		}
	}
	peer.Fetched = true
	list, err := peer.fetchNodes(ctx)
	if err != nil {
		log.Println("failed get list of nodes from", peer.Address, ":", err)
		return
	}
	for _, node := range list {
		err = peer.network.Put(node)
		if err != nil {
			log.Println("failed save node", node.Name, "from", peer.Address, ":", err)
		}
	}
}

func (peer *Peer) fetchConfig(ctx context.Context) (*network.Node, error) {
	url := "http://" + peer.Address + ":" + strconv.Itoa(network.CommunicationPort) + "/"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%d %s", res.StatusCode, res.Status)
	}

	data, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	var node network.Node
	return &node, node.Parse(data)
}

func (peer *Peer) fetchNodes(ctx context.Context) ([]*network.Node, error) {
	url := "http://" + peer.Address + ":" + strconv.Itoa(network.CommunicationPort) + "/rpc/nodes"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%d %s", res.StatusCode, res.Status)
	}

	data, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	var nodes nodeList

	return nodes.Nodes, json.Unmarshal(data, &nodes)
}
