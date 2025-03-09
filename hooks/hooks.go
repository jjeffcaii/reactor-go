package hooks

import (
	"context"
	"sync"

	"github.com/jjeffcaii/reactor-go"
)

//var globalHooks = &Hooks{}

type Hooks[T any] struct {
	mu        sync.RWMutex
	nextDrops []reactor.FnOnDiscard[T]
	errDrops  []reactor.FnOnError
}

func (h *Hooks[T]) OnNextDrop(ctx context.Context, t T) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	for _, fn := range h.nextDrops {
		fn(ctx, t)
	}
}

func (h *Hooks[T]) OnErrorDrop(ctx context.Context, e error) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	for _, fn := range h.errDrops {
		fn(ctx, e)
	}
}

func (h *Hooks[T]) registerOnNextDrop(fn reactor.FnOnDiscard[T]) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.nextDrops = append(h.nextDrops, fn)
}

func (h *Hooks[T]) registerOnErrorDrop(fn reactor.FnOnError) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.errDrops = append(h.errDrops, fn)
}

func Global[T any]() *Hooks[T] {
	panic("todo")
	//return globalHooks
}

func OnNextDrop[T any](fn reactor.FnOnDiscard[T]) {
	//globalHooks.registerOnNextDrop(fn)
}

func OnErrorDrop(fn reactor.FnOnError) {
	//globalHooks.registerOnErrorDrop(fn)
}
