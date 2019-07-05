package mono

import (
	"context"
	"time"

	"github.com/jjeffcaii/reactor-go"
	"github.com/jjeffcaii/reactor-go/scheduler"
)

type wrapper struct {
	raw
}

func (p wrapper) Subscribe(ctx context.Context, options ...rs.SubscriberOption) {
	p.SubscribeWith(ctx, rs.NewSubscriber(options...))
}

func (p wrapper) SwitchIfEmpty(alternative Mono) Mono {
	return wrapper{newMonoSwitchIfEmpty(p, alternative)}
}

func (p wrapper) Filter(f rs.Predicate) Mono {
	return wrapper{newMonoFilter(p, f)}
}

func (p wrapper) Map(t rs.Transformer) Mono {
	return wrapper{newMonoMap(p, t)}
}

func (p wrapper) FlatMap(mapper flatMapper) Mono {
	return wrapper{newMonoFlatMap(p, mapper)}
}

func (p wrapper) SubscribeOn(sc scheduler.Scheduler) Mono {
	return wrapper{newMonoScheduleOn(p, sc)}
}

func (p wrapper) DoOnNext(fn rs.FnOnNext) Mono {
	return wrapper{newMonoPeek(p, peekNext(fn))}
}

func (p wrapper) DoOnError(fn rs.FnOnError) Mono {
	return wrapper{newMonoPeek(p, peekError(fn))}
}

func (p wrapper) DoOnComplete(fn rs.FnOnComplete) Mono {
	return wrapper{newMonoPeek(p, peekComplete(fn))}
}

func (p wrapper) DoOnCancel(fn rs.FnOnCancel) Mono {
	return wrapper{newMonoPeek(p, peekCancel(fn))}
}

func (p wrapper) DoFinally(fn rs.FnOnFinally) Mono {
	return wrapper{newMonoDoFinally(p, fn)}
}

func (p wrapper) DelayElement(delay time.Duration) Mono {
	return wrapper{newMonoDelayElement(p, delay, scheduler.Elastic())}
}

func (p wrapper) Block(ctx context.Context) (v interface{}, err error) {
	cv := make(chan interface{}, 1)
	ce := make(chan error, 1)
	p.Subscribe(ctx,
		rs.OnNext(func(s rs.Subscription, v interface{}) {
			cv <- v
		}),
		rs.OnComplete(func() {
			close(cv)
		}),
		rs.OnError(func(e error) {
			ce <- e
			close(ce)
		}),
	)
	select {
	case v = <-cv:
		close(ce)
	case err = <-ce:
		close(cv)
	}
	return
}
