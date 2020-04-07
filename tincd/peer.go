package tincd

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
	"tinc-web-boot/network"
)

type peerReq struct {
	Node   string
	Subnet string
	Add    bool
}

type peersManager struct {
	lock    sync.RWMutex
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
				peer := pl.Add(ctx, req.Node, req.Subnet)
				wg.Add(1)
				go func() {
					defer wg.Done()
					peer.run(ctx)
				}()
			} else {
				pl.Remove(req.Node)

			}
		}
	}
	wg.Wait()
}

func (pl *peersManager) Add(ctx context.Context, node, subnet string) *Peer {
	ctx, cancel := context.WithCancel(ctx)
	p := &Peer{
		Node:    node,
		Subnet:  subnet,
		stop:    cancel,
		network: pl.network,
	}
	pl.add(p)
	return p
}

func (pl *peersManager) Remove(node string) {
	v, ok := pl.remove(node)
	if ok {
		v.stop()
	}
}

func (pl *peersManager) add(newPeer *Peer) {
	pl.lock.Lock()
	defer pl.lock.Unlock()
	if pl.list == nil {
		pl.list = make(map[string]*Peer)
	}
	pl.list[newPeer.Node] = newPeer
}

func (pl *peersManager) remove(name string) (*Peer, bool) {
	pl.lock.Lock()
	defer pl.lock.Unlock()
	v, ok := pl.list[name]
	delete(pl.list, name)
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
		return cp[i].Node < cp[j].Node
	})
	return cp
}

type Peer struct {
	Node    string
	Subnet  string
	Fetched bool
	stop    func()
	network *network.Network
}

func (peer *Peer) run(ctx context.Context) {
	for {
		node, err := peer.fetchConfig(ctx)
		if err == nil {
			err = peer.network.Put(node)
		}
		if err == nil {
			break
		}
		log.Println("failed get", peer.Node, ":", err)
		select {
		case <-ctx.Done():
			return
		case <-time.After(retryInterval):
		}
	}
	peer.Fetched = true
}

func (peer *Peer) fetchConfig(ctx context.Context) (*network.Node, error) {
	addr := strings.TrimSpace(strings.Split(peer.Subnet, "/")[0])
	url := "http://" + addr + ":" + strconv.Itoa(network.CommunicationPort) + "/"
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
	return &node, node.UnmarshalText(data)
}
