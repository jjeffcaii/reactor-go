package hooks

import (
	"sync"

	"github.com/jjeffcaii/reactor-go"
)

var globalHooks = &Hooks{}

type Hooks struct {
	sync.RWMutex
	nextDrops []reactor.FnOnDiscard
	errDrops  []reactor.FnOnError
}

func (h *Hooks) OnNextDrop(t reactor.Any) {
	h.RLock()
	defer h.RUnlock()
	for _, fn := range h.nextDrops {
		fn(t)
	}
}

func (h *Hooks) OnErrorDrop(e error) {
	h.RLock()
	defer h.RUnlock()
	for _, fn := range h.errDrops {
		fn(e)
	}
}

func (h *Hooks) registerOnNextDrop(fn reactor.FnOnDiscard) {
	h.Lock()
	defer h.Unlock()
	h.nextDrops = append(h.nextDrops, fn)
}

func (h *Hooks) registerOnErrorDrop(fn reactor.FnOnError) {
	h.Lock()
	defer h.Unlock()
	h.errDrops = append(h.errDrops, fn)
}

func Global() *Hooks {
	return globalHooks
}

func OnNextDrop(fn reactor.FnOnDiscard) {
	globalHooks.registerOnNextDrop(fn)
}

func OnErrorDrop(fn reactor.FnOnError) {
	globalHooks.registerOnErrorDrop(fn)
}
