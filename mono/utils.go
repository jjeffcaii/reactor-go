package mono

import (
	"context"
	"time"
)

var empty = wrap(newMonoJust(nil))
var _errJustNilValue = "require non nil value"

func Error(e error) Mono {
	return wrap(newMonoError(e))
}

func ErrorOneshot(e error) Mono {
	return borrowOneshotWrapper(newMonoError(e))
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
		panic(_errJustNilValue)
	}
	return wrap(newMonoJust(v))
}

func JustOneshot(v Any) Mono {
	if v == nil {
		panic(_errJustNilValue)
	}
	return borrowOneshotWrapper(newMonoJust(v))
}

func Create(gen func(ctx context.Context, s Sink)) Mono {
	return wrap(newMonoCreate(gen))
}

func CreateOneshot(gen func(ctx context.Context, s Sink)) Mono {
	return borrowOneshotWrapper(newMonoCreate(gen))
}

func Delay(delay time.Duration) Mono {
	return wrap(newMonoDelay(delay))
}

func CreateProcessor() Processor {
	return wrap(&processor{})
}
