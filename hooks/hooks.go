package hooks

import (
	rs "github.com/jjeffcaii/reactor-go"
)

var globalHooks = &Hooks{}

type Hooks struct {
	nextDrops []rs.FnOnDiscard
	errDrops  []rs.FnOnError
}

func (p *Hooks) OnNextDrop(t interface{}) {
	for _, fn := range p.nextDrops {
		fn(t)
	}
}

func (p *Hooks) OnErrorDrop(e error) {
	for _, fn := range p.errDrops {
		fn(e)
	}
}

func (p *Hooks) registerOnNextDrop(fn rs.FnOnDiscard) {
	p.nextDrops = append(p.nextDrops, fn)
}

func (p *Hooks) registerOnErrorDrop(fn rs.FnOnError) {
	p.errDrops = append(p.errDrops, fn)
}

func Global() *Hooks {
	return globalHooks
}

func OnNextDrop(fn rs.FnOnDiscard) {
	globalHooks.registerOnNextDrop(fn)
}

func OnErrorDrop(fn rs.FnOnError) {
	globalHooks.registerOnErrorDrop(fn)
}
