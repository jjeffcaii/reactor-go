package mono

import (
	"context"
	"time"

	"github.com/jjeffcaii/reactor-go/scheduler"
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

func Zip(first, second Mono, others ...Mono) Mono {
	sources := make([]Mono, len(others)+2)
	sources[0] = first
	sources[1] = second
	for i := 0; i < len(others); i++ {
		sources[2+i] = others[i]
	}
	return wrap(newMonoZip(sources))
}

func ZipAll(all ...Mono) Mono {
	if len(all) < 1 {
		panic("at least one Mono for zip operation")
	}
	return wrap(newMonoZip(all))
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

func NewProcessor(sc scheduler.Scheduler, hook ProcessorFinallyHook) (Mono, Sink, Disposable) {
	p := globalProcessorPool.get()
	p.mu.Lock()
	p.sc = sc
	p.hookOnFinally = hook
	p.mu.Unlock()
	return wrap(p), p, p
}
