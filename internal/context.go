package internal

import (
	"context"

	rs "github.com/jjeffcaii/reactor-go"
)

type Key string

const (
	KeyOnDiscard Key = "onDiscard"
	KeyOnError   Key = "onError"
)

type ContextSupport struct {
	ctx context.Context
}

func (c *ContextSupport) Context() context.Context {
	return c.ctx
}

func NewContextSupport(ctx context.Context) *ContextSupport {
	return &ContextSupport{
		ctx: ctx,
	}
}

func TryDiscard(cs *ContextSupport, v interface{}) {
	fn, ok := cs.ctx.Value(KeyOnDiscard).(rs.FnOnDiscard)
	if !ok {
		return
	}
	fn(v)
}
