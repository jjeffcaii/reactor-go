package rs

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
	Predicate func(v interface{}) bool
	Transformer func(v interface{}) interface{}

	FnOnComplete = func()
	FnOnNext = func(v interface{})
	FnOnCancel = func()
	FnOnSubscribe = func(su Subscription)
	FnOnRequest = func(n int)
	FnOnError = func(e error)
	FnOnFinally = func(s SignalType)
	FnOnDiscard = func(v interface{})
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
