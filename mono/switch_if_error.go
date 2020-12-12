package mono

import (
	"context"

	"github.com/jjeffcaii/reactor-go"
	"github.com/jjeffcaii/reactor-go/internal/subscribers"
	"github.com/pkg/errors"
)

type monoSwitchIfError struct {
	source reactor.RawPublisher
	sw     func(error) Mono
}

func (m monoSwitchIfError) SubscribeWith(ctx context.Context, actual reactor.Subscriber) {
	alternative := func(err error) (pub reactor.RawPublisher) {
		if m.sw == nil {
			pub = newMonoError(errors.New("the SwitchIfError transform is nil"))
			return
		}
		defer func() {
			rec := recover()
			if rec == nil {
				return
			}
			if e, ok := rec.(error); ok {
				pub = newMonoError(errors.WithStack(e))
			} else {
				pub = newMonoError(errors.Errorf("%v", rec))
			}
		}()
		pub = m.sw(err)
		if pub == nil {
			pub = newMonoError(errors.New("the SwitchIfError returns nil Mono"))
		}
		return
	}
	s := subscribers.NewSwitchIfErrorSubscriber(alternative, actual)
	actual.OnSubscribe(ctx, s)
	m.source.SubscribeWith(ctx, s)
}

func newMonoSwitchIfError(source reactor.RawPublisher, sw func(error) Mono) monoSwitchIfError {
	return monoSwitchIfError{
		source: source,
		sw:     sw,
	}
}
