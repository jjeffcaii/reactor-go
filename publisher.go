package rs

import "context"

type RawPublisher interface {
	SubscribeWith(context.Context, Subscriber)
}

type Publisher interface {
	RawPublisher
	Subscribe(context.Context, ...SubscriberOption)
}
