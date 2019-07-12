package mono

import (
	"context"

	"github.com/jjeffcaii/reactor-go"
	"github.com/jjeffcaii/reactor-go/internal"
)

type monoContext struct {
	source Mono
	kv     []interface{}
}

func (p *monoContext) SubscribeWith(ctx context.Context, s rs.Subscriber) {
	for i := 0; i < len(p.kv); i += 2 {
		ctx = context.WithValue(ctx, p.kv[i], p.kv[i+1])
	}
	p.source.SubscribeWith(ctx, internal.NewCoreSubscriber(ctx, s))
}

type monoContextOption func(*monoContext)

func withContextDiscard(fn rs.FnOnDiscard) monoContextOption {
	return func(i *monoContext) {
		i.kv = append(i.kv, internal.KeyOnDiscard, fn)
	}
}

func withContextError(fn rs.FnOnError) monoContextOption {
	return func(i *monoContext) {
		i.kv = append(i.kv, internal.KeyOnError, fn)
	}
}

func newMonoContext(source Mono, options ...monoContextOption) *monoContext {
	mc := &monoContext{
		source: source,
	}
	for _, it := range options {
		it(mc)
	}
	return mc
}
