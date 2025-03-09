package internal

import (
	"context"

	"github.com/jjeffcaii/reactor-go"
)

type Key string

const (
	KeyOnDiscard Key = "onDiscard"
	KeyOnError   Key = "onError"
)

func TryDiscard[T any](ctx context.Context, v T) {
	fn, ok := ctx.Value(KeyOnDiscard).(reactor.FnOnDiscard[T])
	if !ok {
		return
	}
	fn(ctx, v)
}
