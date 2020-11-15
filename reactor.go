package reactor

import "context"

// Item is type of element.
type Item struct {
	V Any
	E error
}

// SignalType is type of terminal signal.
type SignalType int8

func (s SignalType) String() string {
	switch s {
	case SignalTypeComplete:
		return "COMPLETE"
	case SignalTypeDefault:
		return "DEFAULT"
	case SignalTypeCancel:
		return "CANCEL"
	case SignalTypeError:
		return "ERROR"
	default:
		return "UNKNOWN"
	}
}

const (
	SignalTypeDefault SignalType = iota
	SignalTypeComplete
	SignalTypeCancel
	SignalTypeError
)

// Any is an alias of interface{} which means a value of any type.
type Any = interface{}

type (
	Predicate   func(Any) bool
	Transformer func(Any) (Any, error)
)

// A group of action functions.
type (
	FnOnComplete  = func()
	FnOnNext      = func(v Any) error
	FnOnCancel    = func()
	FnOnSubscribe = func(ctx context.Context, su Subscription)
	FnOnRequest   = func(n int)
	FnOnError     = func(e error)
	FnOnFinally   = func(s SignalType)
	FnOnDiscard   = func(v Any)
)

// Disposable is a disposable resource.
type Disposable interface {
	// Dispose dispose current resource.
	Dispose()
	// IsDisposed returns true if it has been disposed.
	IsDisposed() bool
}
