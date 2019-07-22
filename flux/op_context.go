package flux

import (
	"context"

	rs "github.com/jjeffcaii/reactor-go"
	"github.com/jjeffcaii/reactor-go/internal"
)

type fluxContext struct {
	source rs.RawPublisher
	kv     []interface{}
}

func (p *fluxContext) SubscribeWith(ctx context.Context, s rs.Subscriber) {
	for i := 0; i < len(p.kv); i += 2 {
		ctx = context.WithValue(ctx, p.kv[i], p.kv[i+1])
	}
	p.source.SubscribeWith(ctx, internal.NewCoreSubscriber(ctx, s))
}

type fluxContextOption func(*fluxContext)

func withContextDiscard(fn rs.FnOnDiscard) fluxContextOption {
	return func(i *fluxContext) {
		i.kv = append(i.kv, internal.KeyOnDiscard, fn)
	}
}

func withContextError(fn rs.FnOnError) fluxContextOption {
	return func(i *fluxContext) {
		i.kv = append(i.kv, internal.KeyOnError, fn)
	}
}

func newFluxContext(source rs.RawPublisher, options ...fluxContextOption) *fluxContext {
	f := &fluxContext{
		source: source,
	}
	for _, it := range options {
		it(f)
	}
	return f
}
