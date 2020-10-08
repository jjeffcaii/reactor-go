package flux

import (
	"context"

	"github.com/jjeffcaii/reactor-go"
	"github.com/jjeffcaii/reactor-go/internal"
)

type fluxContext struct {
	source reactor.RawPublisher
	kv     []Any
}

func (p *fluxContext) SubscribeWith(ctx context.Context, s reactor.Subscriber) {
	for i := 0; i < len(p.kv); i += 2 {
		ctx = context.WithValue(ctx, p.kv[i], p.kv[i+1])
	}
	p.source.SubscribeWith(ctx, s)
}

type fluxContextOption func(*fluxContext)

func withContextDiscard(fn reactor.FnOnDiscard) fluxContextOption {
	return func(i *fluxContext) {
		i.kv = append(i.kv, internal.KeyOnDiscard, fn)
	}
}

func newFluxContext(source reactor.RawPublisher, options ...fluxContextOption) *fluxContext {
	f := &fluxContext{
		source: source,
	}
	for _, it := range options {
		it(f)
	}
	return f
}
