package hooks

import (
	"sync"

	"github.com/jjeffcaii/reactor-go"
)

var globalHooks = &Hooks{}

type Hooks struct {
	nextDrops []reactor.FnOnDiscard
	errDrops  []reactor.FnOnError
	locker    sync.RWMutex
}

func (p *Hooks) OnNextDrop(t reactor.Any) {
	p.locker.RLock()
	for _, fn := range p.nextDrops {
		fn(t)
	}
	p.locker.RUnlock()
}

func (p *Hooks) OnErrorDrop(e error) {
	p.locker.RLock()
	for _, fn := range p.errDrops {
		fn(e)
	}
	p.locker.RUnlock()
}

func (p *Hooks) registerOnNextDrop(fn reactor.FnOnDiscard) {
	p.locker.Lock()
	p.nextDrops = append(p.nextDrops, fn)
	p.locker.Unlock()
}

func (p *Hooks) registerOnErrorDrop(fn reactor.FnOnError) {
	p.locker.Lock()
	p.errDrops = append(p.errDrops, fn)
	p.locker.Unlock()
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
