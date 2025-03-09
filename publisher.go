//go:generate mockgen -destination=./testdata/mock_publisher.go -package=testdata . RawPublisher,Publisher
package reactor

import "context"

// RawPublisher is the basic low-level Publisher that can be subscribed with a Subscriber.
type RawPublisher[T any] interface {
	// SubscribeWith subscribes current Publisher with a Subscriber.
	SubscribeWith(context.Context, Subscriber[T])
}

// Publisher is th basic type that can be subscribed
type Publisher[T any] interface {
	RawPublisher[T]
	// Subscribe subscribes current Publisher with some options.
	Subscribe(context.Context, ...SubscriberOption[T])
}
