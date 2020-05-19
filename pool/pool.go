package pool

import (
	"context"
	"fmt"
	"github.com/tinc-boot/tincd"
	"github.com/tinc-boot/tincd/network"
	"net"
	"path/filepath"
	"sync"
)

func New(ctx context.Context, configFile, rootDir, tincBin string) (*Pool, error) {
	pool := &Pool{
		tincBin: tincBin,
		rootDir: rootDir,
		nets:    map[string]tincd.Tincd{},
		ctx:     ctx,
	}

	err := pool.Config.LoadFrom(configFile)
	if err != nil {
		return nil, err
	}

	list, err := network.List(rootDir)
	if err != nil {
		return nil, err
	}

	var toStart []*network.Network

	for _, ntw := range list {
		if pool.Config.AutoStart.Has(ntw.Name()) {
			toStart = append(toStart, ntw)
		}
	}

	for _, ntw := range toStart {
		_, err := pool.RunNetwork(ntw)
		if err != nil {
			pool.Stop()
			return nil, err
		}
	}

	return pool, nil
}

type Pool struct {
	tincBin string
	rootDir string
	lock    sync.Mutex
	ctx     context.Context
	nets    map[string]tincd.Tincd
	Config  Config
	events  network.Events
}

func (pool *Pool) Events() *network.Events {
	return &pool.events
}

func (pool *Pool) Find(name string) tincd.Tincd {
	pool.lock.Lock()
	defer pool.lock.Unlock()
	return pool.nets[name]
}

func (pool *Pool) IsRunning(name string) bool {
	instance := pool.Find(name)
	return instance != nil && instance.IsRunning()
}

func (pool *Pool) Network(name string) (*network.Network, error) {
	if !network.IsValidName(name) {
		return nil, fmt.Errorf("invalid name for network")
	}
	return &network.Network{Root: filepath.Join(pool.rootDir, name)}, nil
}

func (pool *Pool) RunNetwork(ntw *network.Network) (tincd.Tincd, error) {
	pool.lock.Lock()
	defer pool.lock.Unlock()
	runningInstance, ok := pool.nets[ntw.Name()]
	if ok {
		return runningInstance, nil
	}
	if pool.nets == nil {
		pool.nets = make(map[string]tincd.Tincd)
	}

	instance, err := tincd.Start(pool.ctx, ntw, false)
	if err != nil {
		return nil, err
	}
	pool.nets[ntw.Name()] = instance
	go func() {
		<-instance.Done()
		pool.lock.Lock()
		delete(pool.nets, ntw.Name())
		pool.lock.Unlock()
	}()
	return instance, nil
}

func (pool *Pool) Create(name string, subnet *net.IPNet) (*network.Network, error) {
	pool.lock.Lock()
	defer pool.lock.Unlock()
	return tincd.CreateNet(filepath.Join(pool.rootDir, name), subnet)
}

func (pool *Pool) Remove(name string) (bool, error) {
	pool.lock.Lock()
	defer pool.lock.Unlock()

	v, ok := pool.nets[name]
	delete(pool.nets, name)

	if ok {
		v.Stop()
		<-v.Done()
		return ok, v.Error()
	}
	return ok, nil
}

func (pool *Pool) Nets() ([]*network.Network, error) {
	return network.List(pool.rootDir)
}

func (pool *Pool) Stop() {
	var wg sync.WaitGroup

	for _, impl := range pool.nets {
		wg.Add(1)
		go func(impl tincd.Tincd) {
			defer wg.Done()
			impl.Stop()
			<-impl.Done()
		}(impl)
	}

	wg.Wait()
}
