//go:generate mockgen -destination=./testdata/mock_subscriber.go -package=testdata . Subscription,Subscriber
package reactor

import (
	"context"
	"math"
)

// RequestInfinite means request items indefinitely.
const RequestInfinite = math.MaxInt32

// Subscription represents a one-to-one lifecycle of a Subscriber subscribing to a Publisher.
type Subscription interface {
	// Request requests the next N items.
	Request(ctx context.Context, n int)
	// Cancel cancels the current lifecycle of subscribing.
	Cancel(ctx context.Context)
}

// Subscriber is the basic type to subscribing the Publisher and consumes the items from upstream.
type Subscriber[T any] interface {
	// OnComplete is successful terminal state.
	OnComplete(context.Context)
	// OnError is failed terminal state.
	OnError(context.Context, error)
	// OnNext is invoked when a data notification sent by the Publisher in response to requests to Subscription.Request(int).
	OnNext(context.Context, T)
	// OnSubscribe is invoked after calling RawPublisher.SubscribeWith(context.Context, Subscriber).
	OnSubscribe(context.Context, Subscription)
}

type subscriber[T any] struct {
	onSubscribe FnOnSubscribe
	onNext      FnOnNext[T]
	onComplete  FnOnComplete
	onError     FnOnError
}

func (p *subscriber[T]) OnComplete(ctx context.Context) {
	if p == nil || p.onComplete == nil {
		return
	}
	p.onComplete(ctx)
}

func (p *subscriber[T]) OnError(ctx context.Context, err error) {
	if p == nil || p.onError == nil {
		return
	}
	p.onError(ctx, err)
}

func (p *subscriber[T]) OnSubscribe(ctx context.Context, s Subscription) {
	if p == nil || p.onSubscribe == nil {
		s.Request(ctx, RequestInfinite)
	} else {
		p.onSubscribe(ctx, s)
	}
}

func (p *subscriber[T]) OnNext(ctx context.Context, i T) {
	if p == nil || p.onNext == nil {
		return
	}
	if err := p.onNext(ctx, i); err != nil {
		p.OnError(ctx, err)
	}
}

// SubscriberOption is used to create a Subscriber easily.
type SubscriberOption[T any] func(*subscriber[T])

// OnNext specified a Subscriber.OnNext action.
func OnNext[T any](onNext FnOnNext[T]) SubscriberOption[T] {
	return func(s *subscriber[T]) {
		s.onNext = onNext
	}
}

// OnComplete specified a Subscriber.OnComplete action.
func OnComplete[T any](onComplete FnOnComplete) SubscriberOption[T] {
	return func(s *subscriber[T]) {
		s.onComplete = onComplete
	}
}

// OnError specified a Subscriber.OnError action.
func OnError[T any](onError FnOnError) SubscriberOption[T] {
	return func(i *subscriber[T]) {
		i.onError = onError
	}
}

// OnSubscribe specified a Subscriber.OnSubscribe action.
func OnSubscribe[T any](onSubscribe FnOnSubscribe) SubscriberOption[T] {
	return func(i *subscriber[T]) {
		i.onSubscribe = onSubscribe
	}
}

// NewSubscriber creates a Subscriber with given options.
func NewSubscriber[T any](opts ...SubscriberOption[T]) Subscriber[T] {
	if len(opts) < 1 {
		return (*subscriber[T])(nil)
	}
	var s subscriber[T]
	for _, opt := range opts {
		opt(&s)
	}
	return &s
}
