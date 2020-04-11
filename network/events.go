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

type eventPeerDiscovered struct {
	lock     sync.RWMutex
	handlers []func(PeerID)
}

func (ev *eventPeerDiscovered) Subscribe(handler func(PeerID)) {
	ev.lock.Lock()
	ev.handlers = append(ev.handlers, handler)
	ev.lock.Unlock()
}
func (ev *eventPeerDiscovered) Emit(payload PeerID) {
	ev.lock.RLock()
	for _, handler := range ev.handlers {
		handler(payload)
	}
	ev.lock.RUnlock()
}

type eventPeerJoined struct {
	lock     sync.RWMutex
	handlers []func(PeerID)
}

func (ev *eventPeerJoined) Subscribe(handler func(PeerID)) {
	ev.lock.Lock()
	ev.handlers = append(ev.handlers, handler)
	ev.lock.Unlock()
}
func (ev *eventPeerJoined) Emit(payload PeerID) {
	ev.lock.RLock()
	for _, handler := range ev.handlers {
		handler(payload)
	}
	ev.lock.RUnlock()
}

type eventPeerLeft struct {
	lock     sync.RWMutex
	handlers []func(PeerID)
}

func (ev *eventPeerLeft) Subscribe(handler func(PeerID)) {
	ev.lock.Lock()
	ev.handlers = append(ev.handlers, handler)
	ev.lock.Unlock()
}
func (ev *eventPeerLeft) Emit(payload PeerID) {
	ev.lock.RLock()
	for _, handler := range ev.handlers {
		handler(payload)
	}
	ev.lock.RUnlock()
}

type Events struct {
	Started        eventStarted
	Stopped        eventStopped
	PeerDiscovered eventPeerDiscovered
	PeerJoined     eventPeerJoined
	PeerLeft       eventPeerLeft
}

func (bus *Events) Sink(sink func(eventName string, payload interface{})) *Events {
	bus.Started.Subscribe(func(payload NetworkID) {
		sink("Started", payload)
	})
	bus.Stopped.Subscribe(func(payload NetworkID) {
		sink("Stopped", payload)
	})
	bus.PeerDiscovered.Subscribe(func(payload PeerID) {
		sink("PeerDiscovered", payload)
	})
	bus.PeerJoined.Subscribe(func(payload PeerID) {
		sink("PeerJoined", payload)
	})
	bus.PeerLeft.Subscribe(func(payload PeerID) {
		sink("PeerLeft", payload)
	})
	return bus
}
func (bus *Events) SubscribeAll(listener interface {
	Started(payload NetworkID)
	Stopped(payload NetworkID)
	PeerDiscovered(payload PeerID)
	PeerJoined(payload PeerID)
	PeerLeft(payload PeerID)
}) {
	bus.Started.Subscribe(listener.Started)
	bus.Stopped.Subscribe(listener.Stopped)
	bus.PeerDiscovered.Subscribe(listener.PeerDiscovered)
	bus.PeerJoined.Subscribe(listener.PeerJoined)
	bus.PeerLeft.Subscribe(listener.PeerLeft)
}
