package flux

import (
	"context"
	"time"
)

type OverflowStrategy int8

const (
	OverflowBuffer OverflowStrategy = iota
	OverflowIgnore
	OverflowError
	OverflowDrop
	OverflowLatest
)

var empty = just(nil)

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
	return wrap(newFluxCreate(c))
}

func Interval(period time.Duration) Flux {
	return wrap(newFluxInterval(period, 0, nil))
}

func just(values []interface{}) Flux {
	return wrap(newSliceFlux(values))
}
