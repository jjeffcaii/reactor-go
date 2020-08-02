package internal

import (
	"sync"

	"github.com/jjeffcaii/reactor-go"
)

type Vector struct {
	v []reactor.Any
	l sync.RWMutex
}

func (p *Vector) Add(v reactor.Any) {
	p.l.Lock()
	p.v = append(p.v, v)
	p.l.Unlock()
}

func (p *Vector) Snapshot() (v []reactor.Any) {
	p.l.RLock()
	if p.v != nil {
		v = p.v[:]
	}
	p.l.RUnlock()
	return
}

func NewVector() *Vector {
	return &Vector{}
}
