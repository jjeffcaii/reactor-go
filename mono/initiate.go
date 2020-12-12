package mono

import (
	"context"
	"fmt"
	"time"

	"github.com/jjeffcaii/reactor-go"
	"github.com/jjeffcaii/reactor-go/scheduler"
)

var empty = wrap(monoEmpty{})
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

func NewProcessor(sc scheduler.Scheduler, hook ProcessorFinallyHook) (Mono, Sink, Disposable) {
	p := globalProcessorPool.get()
	p.mu.Lock()
	p.sc = sc
	p.hookOnFinally = hook
	p.mu.Unlock()
	return wrap(p), p, p
}

func Zip(first Mono, second Mono, rest ...Mono) Mono {
	return wrap(zip(first, second, rest))
}

func ZipCombine(cmb Combinator, sources ...Mono) Mono {
	return wrap(zipCombine(cmb, sources))
}

func ZipOneshot(first Mono, second Mono, rest ...Mono) Mono {
	return borrowOneshotWrapper(zip(first, second, rest))
}

func ZipCombineOneshot(cmb Combinator, sources ...Mono) Mono {
	return borrowOneshotWrapper(zipCombine(cmb, sources))
}

func innerZipCombine(sources []reactor.RawPublisher, cmb Combinator) *monoZip {
	for i := 0; i < len(sources); i++ {
		if sources[i] == nil {
			panic(fmt.Sprintf("the #%d Mono to be zipped is nil!", i))
		}
	}
	return newMonoZip(sources, cmb)
}

func zip(first Mono, second Mono, rest []Mono) *monoZip {
	sources := make([]reactor.RawPublisher, len(rest)+2)
	sources[0] = unpackRawPublisher(first)
	sources[1] = unpackRawPublisher(second)
	for i := 0; i < len(rest); i++ {
		sources[i+2] = unpackRawPublisher(rest[i])
	}
	return innerZipCombine(sources, nil)
}

func zipCombine(cmb Combinator, sources []Mono) *monoZip {
	if len(sources) < 2 {
		panic("ZipCombine need at least two Mono!")
	}
	pubs := make([]reactor.RawPublisher, len(sources))
	for i := 0; i < len(sources); i++ {
		pubs[i] = unpackRawPublisher(sources[i])
	}
	return innerZipCombine(pubs, cmb)
}
