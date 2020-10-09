package mono

import (
	"context"
	"errors"
	"time"
)

var empty = wrap(newMonoJust(nil))
var errJustNilValue = errors.New("require non nil value")

func Error(e error) Mono {
	return getObjFromPool(newMonoError(e))
}

func Empty() Mono {
	return empty
}

func JustOrEmpty(v Any) Mono {
	if v == nil {
		return empty
	}
	return Just(v)
}

func Just(v Any) Mono {
	if v == nil {
		panic(errJustNilValue)
	}
	return getObjFromPool(newMonoJust(v))
}

func Create(gen func(ctx context.Context, s Sink)) Mono {
	return getObjFromPool(newMonoCreate(gen))
}

func Delay(delay time.Duration) Mono {
	return getObjFromPool(newMonoDelay(delay))
}

func CreateProcessor() Processor {
	return getObjFromPool(&processor{})
}
