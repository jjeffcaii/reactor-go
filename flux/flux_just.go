package flux

import (
	"context"
)

func Just(first interface{}, others ...interface{}) Flux {
	return New(func(ctx context.Context, sink Sink) {
		_ = sink.Next(first)
		for _, it := range others {
			_ = sink.Next(it)
		}
		sink.Complete()
	})
}
