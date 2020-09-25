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

func (p *Hooks) OnNextDrop(t reactor.Any) {
	p.RLock()
	defer p.RUnlock()
	for _, fn := range p.nextDrops {
		fn(t)
	}
}

func (p *Hooks) OnErrorDrop(e error) {
	p.RLock()
	defer p.RUnlock()
	for _, fn := range p.errDrops {
		fn(e)
	}
}

func (p *Hooks) registerOnNextDrop(fn reactor.FnOnDiscard) {
	p.Lock()
	defer p.Unlock()
	p.nextDrops = append(p.nextDrops, fn)
}

func (p *Hooks) registerOnErrorDrop(fn reactor.FnOnError) {
	p.Lock()
	defer p.Unlock()
	p.errDrops = append(p.errDrops, fn)
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
