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
	Request(n int)
	// Cancel cancels the current lifecycle of subscribing.
	Cancel()
}

// Subscriber is the basic type to subscribing the Publisher and consumes the items from upstream.
type Subscriber interface {
	// OnComplete is successful terminal state.
	OnComplete()
	// OnError is failed terminal state.
	OnError(error)
	// OnNext is invoked when a data notification sent by the Publisher in response to requests to Subscription.Request(int).
	OnNext(Any)
	// OnSubscribe is invoked after calling RawPublisher.SubscribeWith(context.Context, Subscriber).
	OnSubscribe(context.Context, Subscription)
}

type subscriber struct {
	onSubscribe FnOnSubscribe
	onNext      FnOnNext
	onComplete  FnOnComplete
	onError     FnOnError
}

func (p *subscriber) OnComplete() {
	if p == nil || p.onComplete == nil {
		return
	}
	p.onComplete()
}

func (p *subscriber) OnError(err error) {
	if p == nil || p.onError == nil {
		return
	}
	p.onError(err)
}

func (p *subscriber) OnSubscribe(ctx context.Context, s Subscription) {
	if p == nil || p.onSubscribe == nil {
		s.Request(RequestInfinite)
	} else {
		p.onSubscribe(ctx, s)
	}
}

func (p *subscriber) OnNext(i Any) {
	if p == nil || p.onNext == nil {
		return
	}
	if err := p.onNext(i); err != nil {
		p.OnError(err)
	}
}

// SubscriberOption is used to create a Subscriber easily.
type SubscriberOption func(*subscriber)

// OnNext specified a Subscriber.OnNext action.
func OnNext(onNext FnOnNext) SubscriberOption {
	return func(s *subscriber) {
		s.onNext = onNext
	}
}

// OnComplete specified a Subscriber.OnComplete action.
func OnComplete(onComplete FnOnComplete) SubscriberOption {
	return func(s *subscriber) {
		s.onComplete = onComplete
	}
}

// OnError specified a Subscriber.OnError action.
func OnError(onError FnOnError) SubscriberOption {
	return func(i *subscriber) {
		i.onError = onError
	}
}

// OnSubscribe specified a Subscriber.OnSubscribe action.
func OnSubscribe(onSubscribe FnOnSubscribe) SubscriberOption {
	return func(i *subscriber) {
		i.onSubscribe = onSubscribe
	}
}

// NewSubscriber creates a Subscriber with given options.
func NewSubscriber(opts ...SubscriberOption) Subscriber {
	if len(opts) < 1 {
		return (*subscriber)(nil)
	}
	s := &subscriber{}
	for _, opt := range opts {
		opt(s)
	}
	return s
}
