package mono

import (
	"context"
	"sync"
	"sync/atomic"

	"github.com/jjeffcaii/reactor-go"
	"github.com/jjeffcaii/reactor-go/hooks"
	"github.com/jjeffcaii/reactor-go/tuple"
)

var (
	_ reactor.Subscription = (*zipCoordinator)(nil)
	_ reactor.Subscriber   = (*zipInner)(nil)
)

func newMonoZip(sources []Mono) *monoZip {
	return &monoZip{
		sources: sources,
	}
}

func newZipCoordinator(actual reactor.Subscriber, size int) *zipCoordinator {
	c := &zipCoordinator{
		actual:      actual,
		subscribers: make([]*zipInner, size),
		countdown:   int32(size),
	}
	for i := 0; i < size; i++ {
		c.subscribers[i] = &zipInner{
			coordinator: c,
		}
	}
	return c
}

type zipCoordinator struct {
	sync.Mutex
	actual      reactor.Subscriber
	subscribers []*zipInner
	countdown   int32
	requested   int32
	cancelled   int32
}

func (z *zipCoordinator) Request(n int) {
	if n < 0 {
		panic("illegal request n")
	}
	if n == 0 {
		return
	}
	if !atomic.CompareAndSwapInt32(&z.requested, 0, 1) {
		return
	}
	if atomic.LoadInt32(&z.countdown) == 0 {
		z.actual.OnNext(z.collect())
		z.actual.OnComplete()
	}
}

func (z *zipCoordinator) Cancel() {
	if !atomic.CompareAndSwapInt32(&z.cancelled, 0, 1) {
		return
	}
	var su reactor.Subscription
	for _, subscriber := range z.subscribers {
		subscriber.Lock()
		su = subscriber.su
		subscriber.Unlock()
		if su != nil {
			su.Cancel()
		}
		su = nil
	}
}

func (z *zipCoordinator) signal() {
	// countdown
	if atomic.AddInt32(&z.countdown, -1) != 0 {
		return
	}
	// exit if not requested
	if atomic.LoadInt32(&z.requested) != 1 {
		return
	}
	// finish: collect results
	z.actual.OnNext(z.collect())
	z.actual.OnComplete()
}

func (z *zipCoordinator) collect() tuple.Tuple {
	n := len(z.subscribers)
	items := make([]*reactor.Item, n)
	for i := 0; i < n; i++ {
		items[i] = z.subscribers[i].item
	}
	return tuple.NewTuple(items...)
}

type zipInner struct {
	sync.Mutex
	su          reactor.Subscription
	coordinator *zipCoordinator
	item        *reactor.Item
}

func (z *zipInner) OnComplete() {
	z.Lock()
	defer z.Unlock()
	if z.item == nil {
		z.coordinator.signal()
	}
}

func (z *zipInner) OnError(err error) {
	z.Lock()
	defer z.Unlock()
	if z.item != nil {
		if z.coordinator != nil && z.coordinator.actual != nil{
			z.coordinator.actual.OnError(err)
		}
		hooks.Global().OnErrorDrop(err)
		return
	}
	z.item = &reactor.Item{
		E: err,
	}
	z.coordinator.signal()
}

func (z *zipInner) OnNext(any reactor.Any) {
	z.Lock()
	defer z.Unlock()
	if z.item != nil {
		hooks.Global().OnNextDrop(any)
		return
	}
	z.item = &reactor.Item{
		V: any,
	}
	z.coordinator.signal()
}

func (z *zipInner) OnSubscribe(ctx context.Context, su reactor.Subscription) {
	select {
	case <-ctx.Done():
		z.OnError(reactor.NewContextError(ctx.Err()))
	default:
		var exist bool
		z.Lock()
		exist = z.su != nil
		if !exist {
			z.su = su
		}
		z.Unlock()

		// cancel subscription if parent zip coordinator has been cancelled
		if exist || atomic.LoadInt32(&z.coordinator.cancelled) == 1 {
			su.Cancel()
		} else {
			su.Request(reactor.RequestInfinite)
		}
	}
}

type monoZip struct {
	sources []Mono
}

func (m *monoZip) SubscribeWith(ctx context.Context, sub reactor.Subscriber) {
	select {
	case <-ctx.Done():
		sub.OnError(reactor.NewContextError(ctx.Err()))
	default:
		c := newZipCoordinator(sub, len(m.sources))
		sub.OnSubscribe(ctx, c)

		subscribers := c.subscribers
		for i := 0; i < len(subscribers); i++ {
			m.sources[i].SubscribeWith(ctx, subscribers[i])
		}
	}
}
