package flux

import (
	"context"
	"errors"
	"time"
)

const (
	statCancel   = -1
	statError    = -2
	statComplete = 2
)

var errSubscribeOnce = errors.New("only one subscriber is allow")
var _empty = wrap(newSliceFlux(nil))

func Error(e error) Flux {
	return wrap(newFluxError(e))
}

func Range(start, count int) Flux {
	if count < 1 {
		return _empty
	}
	return wrap(newFluxRange(start, start+count))
}

func Empty() Flux {
	return _empty
}

func Just(values ...Any) Flux {
	if len(values) < 1 {
		return _empty
	}
	return wrap(newSliceFlux(values))
}

func Create(c func(ctx context.Context, sink Sink), options ...CreateOption) Flux {
	return wrap(newFluxCreate(c, options...))
}

func Interval(period time.Duration) Flux {
	return wrap(newFluxInterval(period, 0, nil))
}

func NewUnicastProcessor() Processor {
	return wrap(newUnicastProcessor(BuffSizeSM))
}
