package mono

import (
	"context"
	"github.com/jjeffcaii/reactor-go"
	"github.com/jjeffcaii/reactor-go/internal/subscribers"
)

type monoCreateIfError struct {
	source    reactor.RawPublisher
	v  Any
}

func (m *monoCreateIfError) Parent() reactor.RawPublisher {
	return m.source
}

func (m *monoCreateIfError) SubscribeWith(ctx context.Context, s reactor.Subscriber) {
	m.source.SubscribeWith(ctx, subscribers.NewDoCreateIfErrorSubscriber(s, m.v))
}


func newMonoDoCreateIfError(source reactor.RawPublisher, value Any) *monoCreateIfError {
	return &monoCreateIfError{
		source:    source,
		v: value,
	}
}

