package tincd

import (
	"context"
	"fmt"
	"log"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
	"tinc-web-boot/network"
	"tinc-web-boot/tincd/api/impl/apiclient"
	"tinc-web-boot/tincd/api/impl/apiserver"
	"tinc-web-boot/tincd/runner"
)

const (
	greetInterval = 5 * time.Second
)

type netImpl struct {
	tincBin string
	ctx     context.Context

	done chan struct{}
	stop func()
	lock sync.Mutex

	activePeers sync.Map

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
	go func() {
		defer cancel()
		defer close(done)
		defer impl.events.Stopped.Emit(network.NetworkID{Name: impl.definition.Name()})
		err := impl.run(ctx)
		if err != nil {
			log.Println("failed run network", impl.definition.Name(), ":", err)
		}
		impl.activePeers = sync.Map{}
	}()
}

func (impl *netImpl) Stop() {
	impl.lock.Lock()
	defer impl.lock.Unlock()
	impl.unsafeStop()
}

func (impl *netImpl) Peers() []string {
	var ans []string
	impl.activePeers.Range(func(key, value interface{}) bool {
		ans = append(ans, key.(string))
		return true
	})
	sort.Strings(ans)
	return ans
}

func (impl *netImpl) IsActive(node string) bool {
	_, ok := impl.activePeers.Load(node)
	return ok
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
	self, config, err := impl.Definition().SelfConfig()
	if err != nil {
		return err
	}

	absDir, err := filepath.Abs(impl.definition.Root)
	if err != nil {
		return err
	}

	interfaceName := config.Interface
	if interfaceName == "" { // for darwin
		interfaceName = config.Device[strings.LastIndex(config.Device, "/")+1:]
	}

	ctx, abort := context.WithCancel(global)
	defer abort()

	var wg sync.WaitGroup

	// run tinc service
	wg.Add(1)
	go func() {
		defer wg.Done()
		defer abort()

		for event := range runner.RunTinc(global, impl.tincBin, absDir) {
			if event.Add {
				impl.activePeers.Store(event.Peer.Node, event)
			} else {
				impl.activePeers.Delete(event.Peer.Node)
			}
			log.Printf("%+v", event)
		}

	}()

	// run http API
	wg.Add(1)
	go func() {
		defer wg.Done()
		defer abort()
		for {
			err := apiserver.RunHTTP(ctx, "tcp", self.IP+":"+strconv.Itoa(network.CommunicationPort), impl)
			log.Println(impl.definition.Name(), "api stopped:", err)
			select {
			case <-ctx.Done():
				return
			case <-time.After(time.Second):
				log.Println("trying again...")
			}
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		err := impl.greetEveryone(ctx, *self, greetInterval)
		if err != nil {
			log.Println("greeting failed:", err)
		}
	}()

	// fix: change owner of log file and pid file to process runner
	wg.Add(1)
	go func() {
		defer wg.Done()
		select {
		case <-ctx.Done():
		case <-time.After(2 * time.Second):
			_ = network.ApplyOwnerOfSudoUser(impl.definition.Pidfile())
		}
	}()

	impl.events.Started.Emit(network.NetworkID{Name: impl.definition.Name()})
	impl.activePeers.Store(self.Name, self)
	wg.Wait()
	return ctx.Err()
}

func (impl *netImpl) greetEveryone(ctx context.Context, self network.Node, retryInterval time.Duration) error {
	var wg sync.WaitGroup

	nodes, err := impl.Definition().NodesDefinitions()
	if err != nil {
		return err
	}

	for _, node := range nodes {
		wg.Add(1)
		go func(node network.Node) {
			defer wg.Done()

			var client = apiclient.APIClient{BaseURL: "http://" + node.IP + ":" + strconv.Itoa(network.CommunicationPort)}
			for {
				toImport, err := client.Exchange(ctx, self)
				if err != nil {
					log.Println("greet", node.Name, err)
					goto SLEEP
				}
				for _, node := range toImport {
					err := impl.Definition().Put(&node)
					if err != nil {
						log.Println(node.Name, "import", node.Name, ":", err)
					}
				}
				log.Println("greeted", node.Name)
				break
			SLEEP:
				select {
				case <-ctx.Done():
					return
				case <-time.After(retryInterval):

				}
			}

		}(node)
	}
	wg.Wait()
	return nil
}

func (impl *netImpl) Exchange(remote network.Node) ([]network.Node, error) {
	err := impl.Definition().Put(&remote)
	if err != nil {
		return nil, err
	}
	return impl.Definition().NodesDefinitions()
}
