package mono

import (
	"context"
	"fmt"
	"github.com/jjeffcaii/reactor-go"
	"github.com/jjeffcaii/reactor-go/internal"
	"sync"
	"sync/atomic"
)

type Item struct {
	V Any
	Err error
}
type Items struct {
	It []*Item
}

type monoZip struct {
	value []Mono
}

type zipSubscription struct {
	actual reactor.Subscriber
	zVal   *monoZip
	zCtx   context.Context
	n      int32
}
var _zipSubscriptionPool = sync.Pool{
	New: func() interface{} {
		return new(zipSubscription)
	},
}

func borrowZipSubscription() *zipSubscription {
	return _zipSubscriptionPool.Get().(*zipSubscription)
}

func returnZipSubscription(s *zipSubscription) {
	if s == nil {
		return
	}
	s.actual = nil
	atomic.StoreInt32(&s.n, 0)
	_justSubscriptionPool.Put(s)
}

func newMonoZip(i ...Mono) *monoZip {
	return &monoZip{
		value: i,
	}
}

func (z *zipSubscription) Request(n int) {
	if z == nil || z.actual == nil  {
		return
	}
	if n < 1 {
		returnZipSubscription(z)
		panic(reactor.ErrNegativeRequest)
	}
	if !atomic.CompareAndSwapInt32(&z.n, 0, statComplete) {
		return
	}
	defer returnZipSubscription(z)
	if z.zVal.value == nil {
		z.actual.OnComplete()
		return
	}
	defer func() {
		if err := internal.TryRecoverError(recover()); err != nil {
			z.actual.OnError(err)
		} else {
			z.actual.OnComplete()
		}
	}()
	var i,e = z.combinedMonos(z.zVal.value)
	if e != nil {
		z.actual.OnComplete()
	}else{
		//fmt.Println("combinedMonos",i)
		z.actual.OnNext(i)
	}
}
func (z *zipSubscription) combinedMonos( m []Mono ) (*Items,error) {
	if len(m) < 1 {
		return nil,fmt.Errorf("Failed to compare combined. Mono is empty or less than 1\n\n")
	}
	var (
		wg sync.WaitGroup
		res = make([]*Item,0)
	)
	wg.Add(len(m))
	for _,sub := range m {
		var done int32 = 0
		sub.Subscribe(z.zCtx,reactor.OnNext(func(r reactor.Any) error {
			if atomic.CompareAndSwapInt32(&done,0,1){
				res = append(res,&Item{Err: nil,V: r})
			}
			return nil
		}),reactor.OnError(func(e error) {
			res = append(res,&Item{Err: e,V: nil})
			wg.Done()
		}),reactor.OnComplete(func() {
			if atomic.CompareAndSwapInt32(&done,0,1){
				res = append(res,&Item{Err: nil,V: nil})
			}else{
				atomic.CompareAndSwapInt32(&done,1,0)
			}
			//res = append(res,&Item{Err: nil,V: nil})
			wg.Done()
		}))
	}
	wg.Wait()
	return &Items{It:res},nil
}
func (z *zipSubscription) Cancel() {
	if z == nil  || z.actual == nil {
		return
	}
	atomic.CompareAndSwapInt32(&z.n, 0, statCancel)
}

func (m *monoZip) SubscribeWith(ctx context.Context, sub reactor.Subscriber) {
	select {
	case <-ctx.Done():
		sub.OnError(reactor.ErrSubscribeCancelled)
	default:
		su := borrowZipSubscription()
		su.actual = sub
		su.zVal = m
		su.zCtx = ctx
		sub.OnSubscribe(ctx, su)
	}
}



