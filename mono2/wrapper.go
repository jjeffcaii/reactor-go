package mono

import (
	"context"
	"time"

	"github.com/jjeffcaii/reactor-go"
	subscribers2 "github.com/jjeffcaii/reactor-go/internal/subscribers2"
	"github.com/jjeffcaii/reactor-go/scheduler"
)

type wrapper[T any] struct {
	reactor.RawPublisher[T]
}

func (w wrapper[T]) Subscribe(ctx context.Context, options ...reactor.SubscriberOption[T]) {
	w.SubscribeWith(ctx, reactor.NewSubscriber(options...))
}

func (w wrapper[T]) SwitchIfEmpty(alternative Mono[T]) Mono[T] {
	return wrap[T](newMonoSwitchIfEmpty[T](w.RawPublisher, alternative.Raw()))
}

func (w wrapper[T]) SwitchValueIfEmpty(alternative T) Mono[T] {
	return wrap[T](newMonoDoCreateIfEmpty[T](w.RawPublisher, alternative))
}

func (w wrapper[T]) SwitchIfError(alternativeFunc func(context.Context, error) Mono[T]) Mono[T] {
	return wrap[T](newMonoSwitchIfError[T](w.RawPublisher, alternativeFunc))
}

func (w wrapper[T]) SwitchValueIfError(v T) Mono[T] {
	return wrap[T](newMonoDoCreateIfError[T](w.RawPublisher, v))
}

func (w wrapper[T]) Filter(f reactor.Predicate[T]) Mono[T] {
	return wrap[T](newMonoFilter[T](w.RawPublisher, f))
}

func (w wrapper[T]) Map(t reactor.Transformer[T, T]) Mono[T] {
	return wrap[T](newMonoMap[T, T](w.RawPublisher, t))
}

func (w wrapper[T]) FlatMap(mapper FlatMapper[T, T]) Mono[T] {
	return wrap[T](newMonoFlatMap[T, T](w.RawPublisher, mapper))
}

func (w wrapper[T]) SubscribeOn(sc scheduler.Scheduler) Mono[T] {
	return wrap(newMonoScheduleOn(w.RawPublisher, sc))
}

func (w wrapper[T]) DoOnNext(fn reactor.FnOnNext[T]) Mono[T] {
	return wrap[T](newMonoPeek(w.RawPublisher, peekNext(fn)))
}

func (w wrapper[T]) DoOnError(fn reactor.FnOnError) Mono[T] {
	return wrap[T](newMonoPeek[T](w.RawPublisher, peekError[T](fn)))
}

func (w wrapper[T]) DoOnComplete(fn reactor.FnOnComplete) Mono[T] {
	return wrap[T](newMonoPeek[T](w.RawPublisher, peekComplete[T](fn)))
}

func (w wrapper[T]) DoOnCancel(fn reactor.FnOnCancel) Mono[T] {
	return wrap[T](newMonoPeek[T](w.RawPublisher, peekCancel[T](fn)))
}

//func (w wrapper) DoOnDiscard(fn reactor.FnOnDiscard) Mono {
//	return wrap(newMonoContext(w.RawPublisher, withContextDiscard(fn)))
//}

func (w wrapper[T]) DoFinally(fn reactor.FnOnFinally) Mono[T] {
	return wrap[T](newMonoDoFinally[T](w.RawPublisher, fn))
}

func (w wrapper[T]) DoOnSubscribe(fn reactor.FnOnSubscribe) Mono[T] {
	return wrap[T](newMonoPeek[T](w.RawPublisher, peekSubscribe[T](fn)))
}

//func (w wrapper) DelayElement(delay time.Duration) Mono {
//	return wrap(newMonoDelayElement(w.RawPublisher, delay, scheduler.Parallel()))
//}

func (w wrapper[T]) Timeout(timeout time.Duration) Mono[T] {
	if timeout <= 0 {
		return w
	}
	return wrap[T](newMonoTimeout(w.RawPublisher, timeout))
}

func (w wrapper[T]) Block(ctx context.Context) (val reactor.Null[T], err error) {
	sub := subscribers2.NewBlockSubscriber[T]()

	w.RawPublisher.SubscribeWith(ctx, sub)
	<-sub.Done()

	if err = sub.E; err != nil {
		return
	}

	val.Valid = sub.Valid
	val.V = sub.V

	return
}

//
//func (w wrapper) ZipWith(other Mono) Mono {
//	return w.ZipCombineWith(other, nil)
//}
//
//func (w wrapper) ZipCombineWith(other Mono, cmb Combinator) Mono {
//	publishers := []reactor.RawPublisher{
//		w.RawPublisher,
//		unpackRawPublisher(other),
//	}
//	return wrap(newMonoZip(publishers, cmb, nil))
//}

func (w wrapper[T]) Raw() reactor.RawPublisher[T] {
	return w.RawPublisher
}

func wrap[T any](r reactor.RawPublisher[T]) wrapper[T] {
	return wrapper[T]{
		RawPublisher: r,
	}
}
