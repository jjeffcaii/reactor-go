package reactor

import "context"

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

type (
	Any         = interface{}
	Predicate   func(Any) bool
	Transformer func(Any) (Any, error)

	FnOnComplete  = func()
	FnOnNext      = func(v Any) error
	FnOnCancel    = func()
	FnOnSubscribe = func(ctx context.Context, su Subscription)
	FnOnRequest   = func(n int)
	FnOnError     = func(e error)
	FnOnFinally   = func(s SignalType)
	FnOnSwitchError   = func(v Any)  Any
	FnOnDiscard   = func(v Any)
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
