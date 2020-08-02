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

func TryDiscard(ctx context.Context, v reactor.Any) {
	fn, ok := ctx.Value(KeyOnDiscard).(reactor.FnOnDiscard)
	if !ok {
		return
	}
	fn(v)
}
