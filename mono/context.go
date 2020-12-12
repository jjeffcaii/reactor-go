package mono

import (
	"context"

	"github.com/jjeffcaii/reactor-go"
	"github.com/jjeffcaii/reactor-go/internal"
)

type monoContext struct {
	source reactor.RawPublisher
	kv     []Any
}

func (p *monoContext) SubscribeWith(ctx context.Context, s reactor.Subscriber) {
	for i := 0; i < len(p.kv); i += 2 {
		ctx = context.WithValue(ctx, p.kv[i], p.kv[i+1])
	}
	p.source.SubscribeWith(ctx, s)
}

type monoContextOption func(*monoContext)

func withContextDiscard(fn reactor.FnOnDiscard) monoContextOption {
	return func(i *monoContext) {
		i.kv = append(i.kv, internal.KeyOnDiscard, fn)
	}
}

func newMonoContext(source reactor.RawPublisher, options ...monoContextOption) *monoContext {
	mc := &monoContext{
		source: source,
	}
	for _, it := range options {
		it(mc)
	}
	return mc
}
