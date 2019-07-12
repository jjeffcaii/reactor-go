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

func TryDiscard(ctx context.Context, v interface{}) {
	fn, ok := ctx.Value(KeyOnDiscard).(rs.FnOnDiscard)
	if !ok {
		return
	}
	fn(v)
}
