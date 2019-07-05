package rs

import "context"

type Publisher interface {
	Subscribe(context.Context, ...SubscriberOption)
	SubscribeWith(context.Context, Subscriber)
}
