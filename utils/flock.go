package utils

import (
	"context"
	"sort"
	"sync"
)

type Runner interface {
	Run(ctx context.Context)
}

type runnerWrapper struct {
	stop func()
	done chan struct{}
}

type Flock struct {
	runners map[string]*runnerWrapper
	lock    sync.RWMutex
	pool    sync.WaitGroup
}

func (fl *Flock) Add(ctx context.Context, name string, run Runner) {
	fl.lock.Lock()
	defer fl.lock.Unlock()
	if fl.runners == nil {
		fl.runners = make(map[string]*runnerWrapper)
	}
	old, ok := fl.runners[name]
	if ok {
		old.stop()
	}
	child, cancel := context.WithCancel(ctx)
	wrap := &runnerWrapper{
		stop: cancel,
		done: make(chan struct{}),
	}
	fl.runners[name] = wrap
	fl.pool.Add(1)
	go func() {
		defer fl.pool.Done()
		defer cancel()
		defer close(wrap.done)
		run.Run(child)
		fl.lock.Lock()
		delete(fl.runners, name)
		fl.lock.Unlock()
	}()
}

func (fl *Flock) Remove(name string) <-chan struct{} {
	fl.lock.Lock()
	v, ok := fl.runners[name]
	delete(fl.runners, name)
	fl.lock.Unlock()
	if ok {
		v.stop()
		return v.done
	}
	ch := make(chan struct{})
	close(ch)
	return ch
}

func (fl *Flock) WaitAll() {
	fl.pool.Wait()
}

func (fl *Flock) List() []string {
	fl.lock.RLock()
	var ans = make([]string, 0, len(fl.runners))
	for k := range fl.runners {
		ans = append(ans, k)
	}
	fl.lock.RUnlock()
	sort.Strings(ans)
	return ans
}
