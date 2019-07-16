package rs

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
	FnOnSubscribe = func(su Subscription)
	FnOnRequest = func(n int)
	FnOnError = func(e error)
	FnOnFinally = func(s Signal)
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
