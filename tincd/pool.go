package tincd

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"tinc-web-boot/network"
)

func New(ctx context.Context, storage *network.Storage, tincBin string) (*poolImpl, error) {
	pool := &poolImpl{
		ctx:     ctx,
		tincBin: tincBin,
		storage: storage,
	}

	list, err := storage.List()
	if err != nil {
		return nil, err
	}

	var toStart []*netImpl

	for _, ntw := range list {
		impl, _ := pool.ensure(ntw)
		cfg, err := ntw.Read()
		if err != nil {
			return nil, fmt.Errorf("read config of network %s: %w", ntw.Name(), err)
		}
		if cfg.AutoStart {
			toStart = append(toStart, impl)
		}
	}

	for _, impl := range toStart {
		impl.Start()
	}

	return pool, nil
}

type poolImpl struct {
	tincBin string

	lock sync.Mutex
	ctx  context.Context
	nets map[string]*netImpl

	storage *network.Storage
}

func (pool *poolImpl) Get(name string) (*netImpl, error) {
	nw := pool.storage.Get(name)
	if !nw.IsDefined() {
		return nil, fmt.Errorf("network %s is not defined", name)
	}
	v, _ := pool.ensure(nw)
	return v, nil
}

func (pool *poolImpl) Create(name string) (*netImpl, error) {
	v, created := pool.ensure(pool.storage.Get(name))
	if created {
		return v, v.definition.Configure(pool.ctx, pool.tincBin)
	}
	return v, nil
}

func (pool *poolImpl) Remove(name string) error {
	pool.lock.Lock()
	defer pool.lock.Unlock()

	v, ok := pool.nets[name]
	delete(pool.nets, name)

	if ok {
		v.Stop()
		return v.definition.Destroy()
	}
	return nil
}

func (pool *poolImpl) Nets() []*netImpl {
	pool.lock.Lock()
	var ans = make([]*netImpl, 0, len(pool.nets))
	for _, v := range pool.nets {
		ans = append(ans, v)
	}
	pool.lock.Unlock()
	sort.Slice(ans, func(i, j int) bool {
		return ans[i].definition.Name() < ans[j].definition.Name()
	})
	return ans
}

func (pool *poolImpl) Stop() {
	var wg sync.WaitGroup

	for _, impl := range pool.Nets() {
		wg.Add(1)
		go func(impl *netImpl) {
			defer wg.Done()
			impl.Stop()
		}(impl)
	}

	wg.Wait()
}

func (pool *poolImpl) ensure(netw *network.Network) (*netImpl, bool) {
	pool.lock.Lock()
	defer pool.lock.Unlock()
	if pool.nets == nil {
		pool.nets = make(map[string]*netImpl)
	}

	if v, ok := pool.nets[netw.Name()]; ok {
		return v, false
	}
	v := &netImpl{
		ctx:        pool.ctx,
		definition: netw,
		tincBin:    pool.tincBin,
	}
	pool.nets[netw.Name()] = v
	return v, true
}
