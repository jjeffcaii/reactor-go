package flux

import (
	"context"
	"errors"

	"github.com/jjeffcaii/reactor-go"
	subscribers2 "github.com/jjeffcaii/reactor-go/internal/subscribers2"
	"github.com/jjeffcaii/reactor-go/scheduler"
)

var (
	errRequireChan     = errors.New("require a chan")
	errRequireSlicePtr = errors.New("require a slice point")
	errWrongElemType   = errors.New("wrong element type")
)

var _ Flux[any] = (*wrapper[any])(nil)

type wrapper[T any] struct {
	reactor.RawPublisher[T]
}

func (w wrapper[T]) Subscribe(ctx context.Context, options ...reactor.SubscriberOption[T]) {
	w.SubscribeWith(ctx, reactor.NewSubscriber(options...))
}

//	func (w wrapper[T]) SubscribeWithChan(ctx context.Context, valueChan interface{}, errChan chan<- error) {
//		typ := reflect.TypeOf(valueChan)
//		if typ.Kind() != reflect.Chan || typ.ChanDir()&reflect.SendDir == 0 {
//			panic(errRequireChan)
//		}
//
//		elemType := typ.Elem()
//		value := reflect.ValueOf(valueChan)
//
//		w.Subscribe(
//			ctx,
//			reactor.OnNext(func(a reactor.Any) error {
//				v := reflect.ValueOf(a)
//				if v.Kind() == elemType.Kind() || v.Type().AssignableTo(elemType) {
//					value.Send(v)
//					return nil
//				}
//				return errWrongElemType
//			}),
//			reactor.OnError(func(e error) {
//				errChan <- e
//			}),
//		)
//	}
//
//	func (w wrapper) BlockToSlice(ctx context.Context, slicePtr interface{}) (err error) {
//		if slicePtr == nil {
//			err = errRequireSlicePtr
//			return
//		}
//		typ := reflect.TypeOf(slicePtr)
//		if typ.Kind() != reflect.Ptr || typ.Elem().Kind() != reflect.Slice {
//			err = errRequireSlicePtr
//			return
//		}
//		elemType := typ.Elem().Elem()
//
//		value := reflect.ValueOf(slicePtr).Elem()
//
//		done := make(chan struct{})
//		w.
//			DoFinally(func(s reactor.SignalType) {
//				close(done)
//			}).
//			Subscribe(
//				ctx,
//				reactor.OnNext(func(a reactor.Any) error {
//					v := reflect.ValueOf(a)
//					if v.Kind() == elemType.Kind() || v.Type().AssignableTo(elemType) {
//						value.Set(reflect.Append(value, v))
//						return nil
//					}
//					return errWrongElemType
//				}),
//				reactor.OnError(func(e error) {
//					err = e
//				}),
//			)
//		<-done
//		return
//	}
//
//	func (w wrapper) DelayElement(delay time.Duration) Flux {
//		return wrap(newFluxDelayElement(w.RawPublisher, delay, scheduler.Elastic()))
//	}

func (w wrapper[T]) Take(n int) Flux[T] {
	return wrap[T](newFluxTake[T](w.RawPublisher, n))
}

func (w wrapper[T]) Filter(f reactor.Predicate[T]) Flux[T] {
	return wrap[T](newFluxFilter[T](w.RawPublisher, f))
}

func (w wrapper[T]) Map(t reactor.Transformer[T, T]) Flux[T] {
	return wrap[T](newFluxMap[T, T](w.RawPublisher, t))
}

func (w wrapper[T]) SubscribeOn(sc scheduler.Scheduler) Flux[T] {
	return wrap[T](newFluxSubscribeOn[T](w.RawPublisher, sc))
}

func (w wrapper[T]) DoOnNext(fn reactor.FnOnNext[T]) Flux[T] {
	return wrap[T](newFluxPeek[T](w.RawPublisher, peekNext[T](fn)))
}

func (w wrapper[T]) DoOnComplete(fn reactor.FnOnComplete) Flux[T] {
	return wrap[T](newFluxPeek[T](w.RawPublisher, peekComplete[T](fn)))
}
func (w wrapper[T]) DoOnRequest(fn reactor.FnOnRequest) Flux[T] {
	return wrap[T](newFluxPeek[T](w.RawPublisher, peekRequest[T](fn)))
}

//func (w wrapper) DoOnDiscard(fn reactor.FnOnDiscard) Flux {
//	return wrap(newFluxContext(w.RawPublisher, withContextDiscard(fn)))
//}

func (w wrapper[T]) DoOnCancel(fn reactor.FnOnCancel) Flux[T] {
	return wrap[T](newFluxPeek[T](w.RawPublisher, peekCancel[T](fn)))
}

func (w wrapper[T]) DoOnError(fn reactor.FnOnError) Flux[T] {
	return wrap[T](newFluxPeek[T](w.RawPublisher, peekError[T](fn)))
}

func (w wrapper[T]) DoOnSubscribe(fn reactor.FnOnSubscribe) Flux[T] {
	return wrap[T](newFluxPeek[T](w.RawPublisher, peekSubscribe[T](fn)))
}

func (w wrapper[T]) DoFinally(fn reactor.FnOnFinally) Flux[T] {
	return wrap[T](newFluxFinally[T](w.RawPublisher, fn))
}

//	func (w wrapper) SwitchOnFirst(fn FnSwitchOnFirst) Flux {
//		return wrap(newFluxSwitchOnFirst(w.RawPublisher, fn))
//	}

func (w wrapper[T]) ToChan(ctx context.Context, cap int) (<-chan T, <-chan error) {
	if cap < 1 {
		cap = 1
	}

	var (
		ch  = make(chan T, cap)
		err = make(chan error, 1)
	)

	w.
		DoFinally(func(ctx context.Context, s reactor.SignalType) {
			if s == reactor.SignalTypeCancel {
				err <- reactor.ErrSubscribeCancelled
			}
			close(ch)
			close(err)
		}).
		SubscribeOn(scheduler.Elastic()).
		Subscribe(ctx,
			reactor.OnNext[T](func(ctx context.Context, v T) error {
				ch <- v
				return nil
			}),
			reactor.OnError[T](func(ctx context.Context, e error) {
				err <- e

			}),
		)
	return ch, err
}

func (w wrapper[T]) BlockFirst(ctx context.Context) (val reactor.Null[T], err error) {
	var (
		done  = make(chan struct{})
		vchan = make(chan T, 1)
		echan = make(chan error, 1)
	)

	w.SubscribeWith(ctx, subscribers2.NewBlockFirstSubscriber(done, vchan, echan))

	<-done

	select {
	case v := <-vchan:
		val.Valid, val.V = true, v
	case err = <-echan:
	default:
	}
	return
}

func (w wrapper[T]) BlockLast(ctx context.Context) (val reactor.Null[T], err error) {
	sub := subscribers2.NewBlockLastSubscriber[T]()
	
	w.SubscribeWith(ctx, sub)

	val.V, val.Valid, err = sub.Block()

	return
}

//
//func (w wrapper) Complete() {
//	w.mustProcessor().Complete()
//}
//
//func (w wrapper) Error(e error) {
//	w.mustProcessor().Error(e)
//}
//
//func (w wrapper) Next(v Any) {
//	w.mustProcessor().Next(v)
//}
//
//func (w wrapper) Close() (err error) {
//	closer, ok := w.RawPublisher.(io.Closer)
//	if !ok {
//		return
//	}
//	err = closer.Close()
//	return
//}
//
//func (w wrapper) mustProcessor() rawProcessor {
//	v, ok := w.RawPublisher.(rawProcessor)
//	if !ok {
//		panic(errors.New("require a processor"))
//	}
//	return v
//}

func (w wrapper[T]) Raw() reactor.RawPublisher[T] {
	return w.RawPublisher
}

func wrap[T any](r reactor.RawPublisher[T]) wrapper[T] {
	return wrapper[T]{r}
}
