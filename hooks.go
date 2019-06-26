package rs

import (
	"sync"
)

var hooksPool = sync.Pool{
	New: func() interface{} {
		return new(Hooks)
	},
}

// Hooks is used to handle rx lifecycle events.
type Hooks struct {
	hOnCancel    []FnOnCancel
	hOnRequest   []FnOnRequest
	hOnSubscribe []FnOnSubscribe
	hOnNext      []FnOnNext
	hOnComplete  []FnOnComplete
	hOnError     []FnOnError
	hOnFinally   []FnOnFinally
}

func (p *Hooks) Reset() {
	if p.hOnCancel != nil {
		p.hOnCancel = p.hOnCancel[:0]
	}
	if p.hOnRequest != nil {
		p.hOnRequest = p.hOnRequest[:0]
	}
	if p.hOnSubscribe != nil {
		p.hOnSubscribe = p.hOnSubscribe[:0]
	}
	if p.hOnNext != nil {
		p.hOnNext = p.hOnNext[:0]
	}
	if p.hOnComplete != nil {
		p.hOnComplete = p.hOnComplete[:0]
	}
	if p.hOnError != nil {
		p.hOnError = p.hOnError[:0]
	}
	if p.hOnFinally != nil {
		p.hOnFinally = p.hOnFinally[:0]
	}
}

func (p *Hooks) OnCancel() {
	if p == nil {
		return
	}
	for i, l := 0, len(p.hOnCancel); i < l; i++ {
		p.hOnCancel[i]()
	}
}

func (p *Hooks) OnRequest(n int) {
	if p == nil {
		return
	}
	for i, l := 0, len(p.hOnRequest); i < l; i++ {
		p.hOnRequest[i](n)
	}
}

func (p *Hooks) OnSubscribe(s Subscription) {
	if p == nil {
		return
	}
	for i, l := 0, len(p.hOnSubscribe); i < l; i++ {
		p.hOnSubscribe[l-i-1](s)
	}
}

func (p *Hooks) OnNext(s Subscription, elem interface{}) {
	if p == nil {
		return
	}
	for i, l := 0, len(p.hOnNext); i < l; i++ {
		p.hOnNext[i](s, elem)
	}
}

func (p *Hooks) OnComplete() {
	if p == nil {
		return
	}
	for i, l := 0, len(p.hOnComplete); i < l; i++ {
		p.hOnComplete[i]()
	}
}

func (p *Hooks) OnError(err error) {
	if p == nil {
		return
	}
	for i, l := 0, len(p.hOnError); i < l; i++ {
		p.hOnError[i](err)
	}
}

func (p *Hooks) OnFinally(sig Signal) {
	if p == nil {
		return
	}
	for i, l := 0, len(p.hOnFinally); i < l; i++ {
		p.hOnFinally[l-i-1](sig)
	}
}

func (p *Hooks) DoOnError(fn FnOnError) {
	p.hOnError = append(p.hOnError, fn)
}

func (p *Hooks) DoOnNext(fn FnOnNext) {
	p.hOnNext = append(p.hOnNext, fn)
}

func (p *Hooks) DoOnRequest(fn FnOnRequest) {
	p.hOnRequest = append(p.hOnRequest, fn)
}

func (p *Hooks) DoOnComplete(fn FnOnComplete) {
	p.hOnComplete = append(p.hOnComplete, fn)
}

func (p *Hooks) DoOnCancel(fn FnOnCancel) {
	p.hOnCancel = append(p.hOnCancel, fn)
}

func (p *Hooks) DoOnSubscribe(fn FnOnSubscribe) {
	p.hOnSubscribe = append(p.hOnSubscribe, fn)
}
func (p *Hooks) DoOnFinally(fn FnOnFinally) {
	p.hOnFinally = append(p.hOnFinally, fn)
}

func BorrowHooks() *Hooks {
	return hooksPool.Get().(*Hooks)
}
func ReturnHooks(h *Hooks) {
	h.Reset()
	hooksPool.Put(h)
}
