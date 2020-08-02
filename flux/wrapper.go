package flux

import (
	"context"
	"errors"
	"reflect"
	"time"

	"github.com/jjeffcaii/reactor-go"
	"github.com/jjeffcaii/reactor-go/hooks"
	"github.com/jjeffcaii/reactor-go/scheduler"
)

var (
	errRequireChan     = errors.New("require a chan")
	errRequireSlicePtr = errors.New("require a slice point")
	errWrongElemType   = errors.New("wrong element type")
)

type wrapper struct {
	reactor.RawPublisher
}

func (w wrapper) BlockToChan(ctx context.Context, ch interface{}) (err error) {
	if ch == nil {
		err = errRequireChan
		return
	}
	typ := reflect.TypeOf(ch)
	if typ.Kind() != reflect.Chan || typ.ChanDir()&reflect.SendDir == 0 {
		err = errRequireChan
		return
	}

	elemType := typ.Elem()
	value := reflect.ValueOf(ch)

	done := make(chan struct{})

	w.
		DoFinally(func(s reactor.SignalType) {
			close(done)
		}).
		Subscribe(
			ctx,
			reactor.OnNext(func(a reactor.Any) error {
				v := reflect.ValueOf(a)
				if v.Kind() == elemType.Kind() || v.Type().AssignableTo(elemType) {
					value.Send(v)
					return nil
				}
				return errWrongElemType
			}),
			reactor.OnError(func(e error) {
				err = e
			}),
		)
	<-done
	return
}

func (w wrapper) BlockToSlice(ctx context.Context, slicePtr interface{}) (err error) {
	if slicePtr == nil {
		err = errRequireSlicePtr
		return
	}
	typ := reflect.TypeOf(slicePtr)
	if typ.Kind() != reflect.Ptr || typ.Elem().Kind() != reflect.Slice {
		err = errRequireSlicePtr
		return
	}
	elemType := typ.Elem().Elem()

	value := reflect.ValueOf(slicePtr).Elem()

	done := make(chan struct{})
	w.
		DoFinally(func(s reactor.SignalType) {
			close(done)
		}).
		Subscribe(
			ctx,
			reactor.OnNext(func(a reactor.Any) error {
				v := reflect.ValueOf(a)
				if v.Kind() == elemType.Kind() || v.Type().AssignableTo(elemType) {
					value.Set(reflect.Append(value, v))
					return nil
				}
				return errWrongElemType
			}),
			reactor.OnError(func(e error) {
				err = e
			}),
		)
	<-done
	return
}

func (w wrapper) Subscribe(ctx context.Context, options ...reactor.SubscriberOption) {
	w.SubscribeWith(ctx, reactor.NewSubscriber(options...))
}

func (w wrapper) DelayElement(delay time.Duration) Flux {
	return wrap(newFluxDelayElement(w.RawPublisher, delay, scheduler.Elastic()))
}

func (w wrapper) Take(n int) Flux {
	return wrap(newFluxTake(w.RawPublisher, n))
}

func (w wrapper) Filter(f reactor.Predicate) Flux {
	return wrap(newFluxFilter(w.RawPublisher, f))
}

func (w wrapper) Map(t reactor.Transformer) Flux {
	return wrap(newFluxMap(w.RawPublisher, t))
}

func (w wrapper) SubscribeOn(sc scheduler.Scheduler) Flux {
	return wrap(newFluxSubscribeOn(w.RawPublisher, sc))
}

func (w wrapper) DoOnNext(fn reactor.FnOnNext) Flux {
	return wrap(newFluxPeek(w.RawPublisher, peekNext(fn)))
}

func (w wrapper) DoOnComplete(fn reactor.FnOnComplete) Flux {
	return wrap(newFluxPeek(w.RawPublisher, peekComplete(fn)))
}
func (w wrapper) DoOnRequest(fn reactor.FnOnRequest) Flux {
	return wrap(newFluxPeek(w.RawPublisher, peekRequest(fn)))
}

func (w wrapper) DoOnDiscard(fn reactor.FnOnDiscard) Flux {
	return wrap(newFluxContext(w.RawPublisher, withContextDiscard(fn)))
}

func (w wrapper) DoOnCancel(fn reactor.FnOnCancel) Flux {
	return wrap(newFluxPeek(w.RawPublisher, peekCancel(fn)))
}

func (w wrapper) DoOnError(fn reactor.FnOnError) Flux {
	return wrap(newFluxPeek(w.RawPublisher, peekError(fn)))
}

func (w wrapper) DoFinally(fn reactor.FnOnFinally) Flux {
	return wrap(newFluxFinally(w.RawPublisher, fn))
}

func (w wrapper) DoOnSubscribe(fn reactor.FnOnSubscribe) Flux {
	return wrap(newFluxPeek(w.RawPublisher, peekSubscribe(fn)))
}

func (w wrapper) SwitchOnFirst(fn FnSwitchOnFirst) Flux {
	return wrap(newFluxSwitchOnFirst(w.RawPublisher, fn))
}

func (w wrapper) ToChan(ctx context.Context, cap int) (<-chan Any, <-chan error) {
	if cap < 1 {
		cap = 1
	}
	ch := make(chan Any, cap)
	err := make(chan error, 1)
	w.
		DoFinally(func(s reactor.SignalType) {
			if s == reactor.SignalTypeCancel {
				err <- reactor.ErrSubscribeCancelled
			}
			close(ch)
			close(err)
		}).
		SubscribeOn(scheduler.Elastic()).
		Subscribe(ctx,
			reactor.OnNext(func(v Any) error {
				ch <- v
				return nil
			}),
			reactor.OnError(func(e error) {
				err <- e
			}),
		)
	return ch, err
}

func (w wrapper) BlockFirst(ctx context.Context) (first Any, err error) {
	done := make(chan struct{})
	var su reactor.Subscription
	w.
		DoFinally(func(s reactor.SignalType) {
			close(done)
		}).
		Subscribe(ctx, reactor.OnNext(func(v Any) error {
			first = v
			su.Cancel()
			return nil
		}), reactor.OnSubscribe(func(s reactor.Subscription) {
			su = s
			su.Request(1)
		}), reactor.OnError(func(e error) {
			err = e
		}))
	<-done
	return
}

func (w wrapper) BlockLast(ctx context.Context) (last Any, err error) {
	done := make(chan struct{})
	w.
		DoFinally(func(s reactor.SignalType) {
			close(done)
		}).
		DoOnCancel(func() {
			err = reactor.ErrSubscribeCancelled
		}).
		Subscribe(ctx,
			reactor.OnNext(func(v Any) error {
				if old := last; old != nil {
					hooks.Global().OnNextDrop(old)
				}
				last = v
				return nil
			}),
			reactor.OnError(func(e error) {
				err = e
			}),
		)
	<-done
	return
}

func (w wrapper) Complete() {
	w.mustProcessor().Complete()
}

func (w wrapper) Error(e error) {
	w.mustProcessor().Error(e)
}

func (w wrapper) Next(v Any) {
	w.mustProcessor().Next(v)
}

func (w wrapper) mustProcessor() rawProcessor {
	v, ok := w.RawPublisher.(rawProcessor)
	if !ok {
		panic(errors.New("require a processor"))
	}
	return v
}

func wrap(r reactor.RawPublisher) wrapper {
	return wrapper{r}
}
