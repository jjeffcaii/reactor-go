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
	Predicate func(interface{}) bool
	FnTransform func(interface{}) interface{}

	FnOnComplete = func()
	FnOnNext = func(Subscription, interface{})
	FnOnCancel = func()
	FnOnSubscribe = func(Subscription)
	FnOnRequest = func(int)
	FnOnError = func(error)
	FnOnFinally = func(Signal)
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

	Scheduler interface {
		Do(ctx context.Context, fn func(ctx context.Context))
	}
)

type OpSubscriber func(*Hooks)

func OnRequest(fn FnOnRequest) OpSubscriber {
	return func(h *Hooks) {
		h.DoOnRequest(fn)
	}
}

func OnNext(fn FnOnNext) OpSubscriber {
	return func(h *Hooks) {
		h.DoOnNext(fn)
	}
}

func OnComplete(fn FnOnComplete) OpSubscriber {
	return func(h *Hooks) {
		h.DoOnComplete(fn)
	}
}

func OnSubscribe(fn FnOnSubscribe) OpSubscriber {
	return func(h *Hooks) {
		h.DoOnSubscribe(fn)
	}
}

func OnError(fn FnOnError) OpSubscriber {
	return func(h *Hooks) {
		h.DoOnError(fn)
	}
}
