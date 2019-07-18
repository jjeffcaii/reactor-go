package internal

import "sync"

type Vector struct {
	v []interface{}
	l sync.RWMutex
}

func (p *Vector) Add(v interface{}) {
	p.l.Lock()
	p.v = append(p.v, v)
	p.l.Unlock()
}

func (p *Vector) Snapshot() (v []interface{}) {
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
