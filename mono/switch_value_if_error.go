package mono

import (
	"context"

	"github.com/jjeffcaii/reactor-go"
	"github.com/jjeffcaii/reactor-go/internal/subscribers"
)

type monoSwitchValueIfError struct {
	source reactor.RawPublisher
	v      Any
}

func (m monoSwitchValueIfError) SubscribeWith(ctx context.Context, s reactor.Subscriber) {
	m.source.SubscribeWith(ctx, subscribers.NewSwitchValueIfErrorSubscriber(s, m.v))
}

func newMonoDoCreateIfError(source reactor.RawPublisher, value Any) *monoSwitchValueIfError {
	return &monoSwitchValueIfError{
		source: source,
		v:      value,
	}
}
