package hooks

import (
  "sync"

  rs "github.com/jjeffcaii/reactor-go"
)

var globalHooks = &Hooks{}

type Hooks struct {
  nextDrops []rs.FnOnDiscard
  errDrops  []rs.FnOnError
  locker    sync.RWMutex
}

func (p *Hooks) OnNextDrop(t interface{}) {
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

func (p *Hooks) registerOnNextDrop(fn rs.FnOnDiscard) {
  p.locker.Lock()
  p.nextDrops = append(p.nextDrops, fn)
  p.locker.Unlock()
}

func (p *Hooks) registerOnErrorDrop(fn rs.FnOnError) {
  p.locker.Lock()
  p.errDrops = append(p.errDrops, fn)
  p.locker.Unlock()
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
