package network

import "sync"

type eventStarted struct {
	lock     sync.RWMutex
	handlers []func(NetworkID)
}

func (ev *eventStarted) Subscribe(handler func(NetworkID)) {
	ev.lock.Lock()
	ev.handlers = append(ev.handlers, handler)
	ev.lock.Unlock()
}
func (ev *eventStarted) Emit(payload NetworkID) {
	ev.lock.RLock()
	for _, handler := range ev.handlers {
		handler(payload)
	}
	ev.lock.RUnlock()
}

type eventStopped struct {
	lock     sync.RWMutex
	handlers []func(NetworkID)
}

func (ev *eventStopped) Subscribe(handler func(NetworkID)) {
	ev.lock.Lock()
	ev.handlers = append(ev.handlers, handler)
	ev.lock.Unlock()
}
func (ev *eventStopped) Emit(payload NetworkID) {
	ev.lock.RLock()
	for _, handler := range ev.handlers {
		handler(payload)
	}
	ev.lock.RUnlock()
}

type Events struct {
	Started eventStarted
	Stopped eventStopped
}

func (bus *Events) Sink(sink func(eventName string, payload interface{})) *Events {
	bus.Started.Subscribe(func(payload NetworkID) {
		sink("Started", payload)
	})
	bus.Stopped.Subscribe(func(payload NetworkID) {
		sink("Stopped", payload)
	})
	return bus
}
func (bus *Events) SubscribeAll(listener interface {
	Started(payload NetworkID)
	Stopped(payload NetworkID)
}) {
	bus.Started.Subscribe(listener.Started)
	bus.Stopped.Subscribe(listener.Stopped)
}
