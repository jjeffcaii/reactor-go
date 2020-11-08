package reactor

import "context"

// RawPublisher is the basic low-level Publisher that can be subscribed with a Subscriber.
type RawPublisher interface {
	// SubscribeWith subscribes current Publisher with a Subscriber.
	SubscribeWith(context.Context, Subscriber)
}

// Publisher is th basic type that can be subscribed
type Publisher interface {
	RawPublisher
	// Subscribe subscribes current Publisher with some options.
	Subscribe(context.Context, ...SubscriberOption)
}
