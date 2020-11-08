package reactor

import (
	"context"
	"math"
)

// RequestInfinite means request items indefinitely.
const RequestInfinite = math.MaxInt32

var emptySubscriber = &subscriber{}

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
	fnOnSubscribe FnOnSubscribe
	fnOnNext      FnOnNext
	fnOnComplete  FnOnComplete
	fnOnError     FnOnError
}

func (p *subscriber) OnComplete() {
	if p == nil || p.fnOnComplete == nil {
		return
	}
	p.fnOnComplete()
}

func (p *subscriber) OnError(err error) {
	if p == nil || p.fnOnError == nil {
		return
	}
	p.fnOnError(err)
}

func (p *subscriber) OnSubscribe(ctx context.Context, s Subscription) {
	if p == nil || p.fnOnSubscribe == nil {
		s.Request(RequestInfinite)
	} else {
		p.fnOnSubscribe(ctx, s)
	}
}

func (p *subscriber) OnNext(i Any) {
	if p.fnOnNext == nil {
		return
	}
	if err := p.fnOnNext(i); err != nil {
		p.OnError(err)
	}
}

// SubscriberOption is used to create a Subscriber easily.
type SubscriberOption func(*subscriber)

// OnNext specified a Subscriber.OnNext action.
func OnNext(onNext FnOnNext) SubscriberOption {
	return func(s *subscriber) {
		s.fnOnNext = onNext
	}
}

// OnComplete specified a Subscriber.OnComplete action.
func OnComplete(onComplete FnOnComplete) SubscriberOption {
	return func(s *subscriber) {
		s.fnOnComplete = onComplete
	}
}

// OnError specified a Subscriber.OnError action.
func OnError(onError FnOnError) SubscriberOption {
	return func(i *subscriber) {
		i.fnOnError = onError
	}
}

// OnSubscribe specified a Subscriber.OnSubscribe action.
func OnSubscribe(onSubscribe FnOnSubscribe) SubscriberOption {
	return func(i *subscriber) {
		i.fnOnSubscribe = onSubscribe
	}
}

// NewSubscriber creates a Subscriber with given options.
func NewSubscriber(opts ...SubscriberOption) Subscriber {
	if len(opts) < 1 {
		return emptySubscriber
	}
	s := &subscriber{}
	for _, opt := range opts {
		opt(s)
	}
	return s
}
