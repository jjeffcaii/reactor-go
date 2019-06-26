package rs

import (
	"context"
)

type Signal int8

const (
	SignalDefault Signal = iota
	SignalComplete
	SignalCancel
	SignalError
)

type (
	FnOnComplete = func()
	FnOnNext = func(s Subscription, v interface{})
	FnOnCancel = func()
	FnOnSubscribe = func(s Subscription)
	FnOnRequest = func(n int)
	FnOnError = func(err error)
	FnOnFinally = func(g Signal)
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
		OnSubscribe(s Subscription)
		OnNext(s Subscription, v interface{})
		OnComplete()
		OnError(err error)
	}

	Subscription interface {
		Request(n int)
		Cancel()
	}

	Processor interface {
		Publisher
		Subscriber
	}

	FluxSink interface {
		Next(v interface{}) error
		Error(e error)
		Complete()
	}

	MonoSink interface {
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
		Filter(fn FnFilter) Flux
		Map(fn FnTransform) Flux
		SubscribeOn(s Scheduler) Flux
		PublishOn(s Scheduler) Flux
	}
)
