package reactor

import (
	"context"
)

var _signalTypeNames = []string{
	SignalTypeDefault:  "DEFAULT",
	SignalTypeComplete: "COMPLETE",
	SignalTypeCancel:   "CANCEL",
	SignalTypeError:    "ERROR",
}

const (
	SignalTypeDefault SignalType = iota
	SignalTypeComplete
	SignalTypeCancel
	SignalTypeError
)

// SignalType is type of terminal signal.
type SignalType int8

func (s SignalType) String() string {
	return _signalTypeNames[s]
}

type Null[T any] struct {
	V     T
	Valid bool
}

// Item is type of element.
type Item[T any] struct {
	Valid bool
	V     T
	E     error
}

// Disposable is a disposable resource.
type Disposable interface {
	// Dispose dispose current resource.
	Dispose()
}

type (
	Predicate[T any]      func(ctx context.Context, next T) bool
	Transformer[A, B any] func(ctx context.Context, next A) (B, error)
)

// A group of action functions.
type (
	FnOnComplete       func(context.Context)
	FnOnNext[T any]    func(ctx context.Context, item T) error
	FnOnCancel         func(context.Context)
	FnOnSubscribe      func(ctx context.Context, su Subscription)
	FnOnRequest        func(ctx context.Context, n int)
	FnOnError          func(ctx context.Context, err error)
	FnOnFinally        func(ctx context.Context, s SignalType)
	FnOnDiscard[T any] func(ctx context.Context, v T)
)
