package mono

import (
	"context"
	stdErrors "errors"

	"github.com/pkg/errors"

	"github.com/jjeffcaii/reactor-go"
	subscribers2 "github.com/jjeffcaii/reactor-go/internal/subscribers2"
)

var (
	errNilTransformer = stdErrors.New("the SwitchIfError transform is nil")
	errNilReceiver    = stdErrors.New("the SwitchIfError receiver is nil")
)

type monoSwitchIfError[T any] struct {
	source      reactor.RawPublisher[T]
	alternative func(context.Context, error) Mono[T]
}

func newMonoSwitchIfError[T any](source reactor.RawPublisher[T], alternative func(context.Context, error) Mono[T]) monoSwitchIfError[T] {
	return monoSwitchIfError[T]{
		source:      source,
		alternative: alternative,
	}
}

func (m monoSwitchIfError[T]) SubscribeWith(ctx context.Context, actual reactor.Subscriber[T]) {
	alternative := func(ctx context.Context, err error) (pub reactor.RawPublisher[T]) {
		if m.alternative == nil {
			pub = newMonoError[T](errNilTransformer)
			return
		}
		defer func() {
			rec := recover()
			if rec == nil {
				return
			}
			if e, ok := rec.(error); ok {
				pub = newMonoError[T](errors.WithStack(e))
			} else {
				pub = newMonoError[T](errors.Errorf("%v", rec))
			}
		}()
		pub = m.alternative(ctx, err)
		if pub == nil {
			pub = newMonoError[T](errNilReceiver)
		}
		return
	}
	sub := subscribers2.NewSwitchIfErrorSubscriber[T](alternative, actual)
	actual.OnSubscribe(ctx, sub)
	m.source.SubscribeWith(ctx, sub)
}
