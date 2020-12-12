package mono

import (
	"context"
	"sync"
	"sync/atomic"

	"github.com/jjeffcaii/reactor-go"
	"github.com/jjeffcaii/reactor-go/hooks"
	"github.com/jjeffcaii/reactor-go/tuple"
	"github.com/pkg/errors"
)

var (
	_ reactor.Subscription = (*zipCoordinator)(nil)
	_ reactor.Subscriber   = (*innerZip)(nil)
)

func newMonoZip(sources []reactor.RawPublisher, cmb Combinator) *monoZip {
	return &monoZip{
		sources: sources,
		cmb:     cmb,
	}
}

func newZipCoordinator(actual reactor.Subscriber, size int, cmb Combinator) *zipCoordinator {
	c := &zipCoordinator{
		actual:      actual,
		subscribers: make([]*innerZip, size),
		countdown:   int32(size),
		cmb:         cmb,
	}
	for i := 0; i < size; i++ {
		c.subscribers[i] = &innerZip{
			coordinator: c,
		}
	}
	return c
}

type zipCoordinator struct {
	sync.Mutex
	actual      reactor.Subscriber
	subscribers []*innerZip
	countdown   int32
	requested   int32
	cancelled   int32
	cmb         Combinator
}

func (zc *zipCoordinator) Request(n int) {
	if n < 0 {
		panic("illegal request n")
	}
	if n == 0 {
		return
	}
	if !atomic.CompareAndSwapInt32(&zc.requested, 0, 1) {
		return
	}
	if atomic.LoadInt32(&zc.countdown) == 0 {
		zc.collect()
	}
}

func (zc *zipCoordinator) Cancel() {
	if !atomic.CompareAndSwapInt32(&zc.cancelled, 0, 1) {
		return
	}
	var su reactor.Subscription
	for _, subscriber := range zc.subscribers {
		subscriber.Lock()
		su = subscriber.su
		subscriber.Unlock()
		if su != nil {
			su.Cancel()
		}
		su = nil
	}
}

func (zc *zipCoordinator) signal() {
	// countdown
	if atomic.AddInt32(&zc.countdown, -1) != 0 {
		return
	}
	// exit if not requested
	if atomic.LoadInt32(&zc.requested) != 1 {
		return
	}
	// finish: collect results
	zc.collect()
}

func (zc *zipCoordinator) collect() {
	n := len(zc.subscribers)
	items := make([]*reactor.Item, n)
	var cur *reactor.Item
	for i := 0; i < n; i++ {
		cur = zc.subscribers[i].item
		if cur == nil {
			continue
		}
		items[i] = cur
	}

	res, err := zc.combine(items)
	if err != nil {
		zc.actual.OnError(err)
		return
	}
	if res != nil {
		zc.actual.OnNext(res)
	}
	zc.actual.OnComplete()
}

func (zc *zipCoordinator) combine(values []*reactor.Item) (result reactor.Any, err error) {
	defer func() {
		rec := recover()
		if rec == nil {
			return
		}
		// TODO: drop value and error
		if e, ok := rec.(error); ok {
			err = errors.WithStack(e)
		} else {
			err = errors.Errorf("%v", rec)
		}
	}()

	if zc.cmb == nil {
		result = tuple.NewTuple(values...)
	} else {
		result, err = zc.cmb(values...)
	}
	return
}

type innerZip struct {
	sync.Mutex
	su          reactor.Subscription
	coordinator *zipCoordinator
	item        *reactor.Item
}

func (iz *innerZip) OnComplete() {
	iz.Lock()
	defer iz.Unlock()
	if iz.item == nil {
		iz.coordinator.signal()
	}
}

func (iz *innerZip) OnError(err error) {
	iz.Lock()
	defer iz.Unlock()
	if iz.item != nil {
		hooks.Global().OnErrorDrop(err)
		return
	}
	iz.item = &reactor.Item{
		E: err,
	}
	iz.coordinator.signal()
}

func (iz *innerZip) OnNext(any reactor.Any) {
	iz.Lock()
	defer iz.Unlock()
	if iz.item != nil {
		hooks.Global().OnNextDrop(any)
		return
	}
	iz.item = &reactor.Item{
		V: any,
	}
	iz.coordinator.signal()
}

func (iz *innerZip) OnSubscribe(ctx context.Context, su reactor.Subscription) {
	select {
	case <-ctx.Done():
		iz.OnError(reactor.NewContextError(ctx.Err()))
	default:
		var exist bool
		iz.Lock()
		exist = iz.su != nil
		if !exist {
			iz.su = su
		}
		iz.Unlock()

		// cancel subscription if parent zip coordinator has been cancelled
		if exist || atomic.LoadInt32(&iz.coordinator.cancelled) == 1 {
			su.Cancel()
		} else {
			su.Request(reactor.RequestInfinite)
		}
	}
}

type monoZip struct {
	sources []reactor.RawPublisher
	cmb     Combinator
}

func (m *monoZip) SubscribeWith(ctx context.Context, sub reactor.Subscriber) {
	select {
	case <-ctx.Done():
		sub.OnError(reactor.NewContextError(ctx.Err()))
	default:
		c := newZipCoordinator(sub, len(m.sources), m.cmb)
		sub.OnSubscribe(ctx, c)

		subscribers := c.subscribers
		for i := 0; i < len(subscribers); i++ {
			m.sources[i].SubscribeWith(ctx, subscribers[i])
		}
	}
}
