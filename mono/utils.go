package mono

import (
	"context"
	"errors"
	"time"

	"github.com/jjeffcaii/reactor-go/scheduler"
)

const (
	statCancel   = -1
	statError    = -2
	statComplete = 2
)

var empty = wrap(newMonoJust(nil))
var errJustNilValue = errors.New("require non nil value")

func Error(e error) Mono {
	return wrap(newMonoError(e))
}

func Empty() Mono {
	return empty
}

func JustOrEmpty(v interface{}) Mono {
	if v == nil {
		return empty
	}
	return Just(v)
}

func Just(v interface{}) Mono {
	if v == nil {
		panic(errJustNilValue)
	}
	return wrap(newMonoJust(v))
}

func Create(gen func(ctx context.Context, s Sink)) Mono {
	return wrap(newMonoCreate(gen))
}

func Delay(delay time.Duration) Mono {
	return wrap(newMonoDelay(delay, scheduler.Elastic()))
}

func CreateProcessor() Processor {
	return wrap(newProcessor())
}
