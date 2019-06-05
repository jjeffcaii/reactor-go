package rs

import "context"

type Signal int8

const (
	SignalDefault Signal = iota
	SignalComplete
	SignalCancel
	SignalError
)

type (
	FnOnComplete = func(ctx context.Context)
	FnOnNext = func(ctx context.Context, s Subscription, v interface{})
	FnOnCancel = func(ctx context.Context)
	FnOnSubscribe = func(ctx context.Context, s Subscription)
	FnOnRequest = func(ctx context.Context, n int)
	FnOnError = func(ctx context.Context, err error)
	FnOnFinally = func(ctx context.Context, g Signal)
)

type (
	// Disposable is a disposable resource.
	Disposable interface {
		// Dispose dispose current resource.
		Dispose()
		// IsDisposed returns true if it has been disposed.
		IsDisposed() bool
	}

	Publisher interface {
		Subscribe(ctx context.Context, opts ...OpSubscriber) Disposable
	}

	Subscriber interface {
		OnSubscribe(ctx context.Context, s Subscription)
		OnNext(ctx context.Context, s Subscription, v interface{})
		OnComplete(ctx context.Context)
		OnError(ctx context.Context, err error)
	}

	Subscription interface {
		Request(n int)
		Cancel()
	}

	Processor interface {
		Publisher
		Subscriber
	}

	Producer interface {
		Next(v interface{}) error
		Error(e error)
		Complete()
	}

	MonoProducer interface {
		Success(v interface{})
		Error(e error)
	}

	Scheduler interface {
		Do(ctx context.Context, fn func(ctx context.Context))
	}

	Mono interface {
		Publisher
		Map(fn FnTransform) Mono
		SubscribeOn(s Scheduler) Mono
		PublishOn(s Scheduler) Mono
	}

	Flux interface {
		Publisher
		Map(fn FnTransform) Flux
		SubscribeOn(s Scheduler) Flux
		PublishOn(s Scheduler) Flux
	}
)
