package mono

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jjeffcaii/reactor-go/scheduler"
)

var empty = wrap(newMonoJust(nil))
var errJustNilValue = errors.New("require non nil value")

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

func Create(gen func(context.Context, Sink)) Mono {
	return wrap(newMonoCreate(gen))
}

func Delay(delay time.Duration) Mono {
	return wrap(newMonoDelay(delay, scheduler.Elastic()))
}

func CreateProcessor() Processor {
	return wrapProcessor(&processor{})
}

func tryRecoverError(re interface{}) error {
	if re == nil {
		return nil
	}
	switch e := re.(type) {
	case error:
		return e
	case string:
		return errors.New(e)
	default:
		return fmt.Errorf("%s", e)
	}
}
