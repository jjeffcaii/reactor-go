package flux

import "context"

type OverflowStrategy int8

const (
	OverflowBuffer OverflowStrategy = iota
	OverflowIgnore
	OverflowError
	OverflowDrop
	OverflowLatest
)

var empty Flux = just(nil)

func Empty() Flux {
	return empty
}

func Just(values ...interface{}) Flux {
	if len(values) < 1 {
		return empty
	}
	return just(values)
}

func Create(c func(ctx context.Context, sink Sink)) Flux {
	return wrapper{newFluxCreate(c)}
}

func just(values []interface{}) Flux {
	return wrapper{newSliceFlux(values)}
}
