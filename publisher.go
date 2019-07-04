package rs

import "context"

type Publisher interface {
	Subscribe(context.Context, ...SubscriberOption)
	SubscribeRaw(context.Context, Subscriber)
}
