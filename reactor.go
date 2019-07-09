package rs

import "errors"

var ErrCancelled = errors.New("subscriber has been cancelled")

type Signal int8

func (s Signal) String() string {
	switch s {
	case SignalComplete:
		return "COMPLETE"
	case SignalDefault:
		return "DEFAULT"
	case SignalCancel:
		return "CANCEL"
	case SignalError:
		return "ERROR"
	default:
		return "UNKNOWN"
	}
}

const (
	SignalDefault Signal = iota
	SignalComplete
	SignalCancel
	SignalError
)

type (
	Predicate func(interface{}) bool
	Transformer func(interface{}) interface{}

	FnOnComplete = func()
	FnOnNext = func(v interface{})
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
)
