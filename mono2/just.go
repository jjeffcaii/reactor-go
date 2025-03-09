package mono

import (
	"context"

	uatomic "go.uber.org/atomic"

	"github.com/jjeffcaii/reactor-go"
)

// Just creates a Mono from an item directly.
func Just[T any](item T) Mono[T] {
	return wrap[T](newMonoJust[T](item))
}

type justSubscription[T any] struct {
	parent    *monoJust[T]
	actual    reactor.Subscriber[T]
	completed uatomic.Bool
}

func (j *justSubscription[T]) Request(ctx context.Context, n int) {
	if n < 1 {
		return
	}

	if !j.completed.CompareAndSwap(false, true) {
		return
	}

	if j.parent.valid {
		j.actual.OnNext(ctx, j.parent.value)
	}
}

func (j *justSubscription[T]) Cancel(ctx context.Context) {
	if !j.completed.CompareAndSwap(false, true) {
		return
	}
	j.actual.OnError(ctx, context.Canceled)
}

type monoJust[T any] struct {
	valid bool
	value T
}

func newMonoJust[T any](v T) *monoJust[T] {
	return &monoJust[T]{
		valid: true,
		value: v,
	}
}

func (m *monoJust[T]) SubscribeWith(ctx context.Context, actual reactor.Subscriber[T]) {
	select {
	case <-ctx.Done():
		actual.OnError(ctx, ctx.Err())
	default:
		su := &justSubscription[T]{
			actual: actual,
			parent: m,
		}
		actual.OnSubscribe(ctx, su)
	}
}
