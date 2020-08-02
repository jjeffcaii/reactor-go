package flux_test

import (
	"context"
	"errors"
	"fmt"
	"log"
	"testing"
	"time"

	"github.com/jjeffcaii/reactor-go"
	"github.com/jjeffcaii/reactor-go/flux"
	"github.com/jjeffcaii/reactor-go/scheduler"
	"github.com/stretchr/testify/assert"
)

var fakeErr = errors.New("fake error")

func Example() {
	gen := func(ctx context.Context, sink flux.Sink) {
		for i := 0; i < 10; i++ {
			v := i
			sink.Next(v)
		}
		sink.Complete()
	}
	done := make(chan struct{})

	var su reactor.Subscription
	flux.Create(gen).
		Filter(func(i Any) bool {
			return i.(int)%2 == 0
		}).
		Map(func(i interface{}) (Any, error) {
			return fmt.Sprintf("#HELLO_%04d", i.(int)), nil
		}).
		SubscribeOn(scheduler.Elastic()).
		Subscribe(context.Background(),
			reactor.OnSubscribe(func(s reactor.Subscription) {
				su = s
				s.Request(1)
			}),
			reactor.OnNext(func(v Any) error {
				fmt.Println("next:", v)
				su.Request(1)
				return nil
			}),
			reactor.OnComplete(func() {
				close(done)
			}),
		)
	<-done
}

var testData = []int{1, 2, 3, 4, 5}

func TestSuite(t *testing.T) {
	var inputs []Any
	for _, value := range testData {
		inputs = append(inputs, value)
	}
	all := make(map[string]func() flux.Flux)
	var instJust = flux.Just(inputs...)
	var instCreate = flux.Create(func(ctx context.Context, sink flux.Sink) {
		for _, it := range testData {
			sink.Next(it)
		}
		sink.Complete()
	})
	var instRange = flux.Range(1, 5)
	all["Just"] = func() flux.Flux {
		return instJust
	}
	all["Create"] = func() flux.Flux {
		return instCreate
	}
	all["Range"] = func() flux.Flux {
		return instRange
	}
	all["Unicast"] = func() flux.Flux {
		vv := flux.NewUnicastProcessor()
		go func() {
			for _, it := range testData {
				vv.Next(it)
			}
			vv.Complete()
		}()
		time.Sleep(100 * time.Millisecond)
		return vv
	}
	for k, gen := range all {
		t.Run(fmt.Sprintf("%s_Request", k), func(t *testing.T) {
			testRequest(gen(), t)
		})
		t.Run(fmt.Sprintf("%s_Peek", k), func(t *testing.T) {
			testPeek(gen(), t)
		})
		t.Run(fmt.Sprintf("%s_Discard", k), func(t *testing.T) {
			testDiscard(gen(), t)
		})
		t.Run(fmt.Sprintf("%s_FilterRequest", k), func(t *testing.T) {
			testFilterRequest(gen(), t)
		})
		t.Run(fmt.Sprintf("%s_BlockLast", k), func(t *testing.T) {
			testBlockLast(gen(), t)
		})
		t.Run(fmt.Sprintf("%s_BlockFirst", k), func(t *testing.T) {
			testBlockFirst(gen(), t)
		})
		// TODO: fix onSubscribe test
		//t.Run(fmt.Sprintf("%s_DoSubscribeOn", k), func(t *testing.T) {
		//	testDoOnSubscribe(gen(), t)
		//})
	}
}

func testDoOnSubscribe(f flux.Flux, t *testing.T) {
	var su reactor.Subscription
	var got int
	f.
		DoOnSubscribe(func(s reactor.Subscription) {
			su = s
			su.Request(1)
		}).
		DoOnNext(func(v interface{}) error {
			log.Println("next:", v)
			got++
			su.Request(1)
			return nil
		}).
		Subscribe(context.Background())
	assert.Equal(t, len(testData), got, "bad amount")
}

func testBlockLast(f flux.Flux, t *testing.T) {
	last, err := f.BlockLast(context.Background())
	assert.NoError(t, err, "block last failed")
	assert.Equal(t, testData[len(testData)-1], last, "value doesn't match")
}

func testFilterRequest(f flux.Flux, t *testing.T) {
	var s reactor.Subscription
	var totals, discards, nexts, requests, filter int
	done := make(chan struct{})
	f.
		DoFinally(func(s reactor.SignalType) {
			assert.Equal(t, reactor.SignalTypeComplete, s, "bad signal")
			close(done)
		}).
		Filter(func(i interface{}) (ok bool) {
			totals++
			ok = i.(int)&1 == 0
			if ok {
				filter++
			}
			return
		}).
		DoOnDiscard(func(v interface{}) {
			discards++
		}).
		DoOnNext(func(v interface{}) error {
			nexts++
			s.Request(1)
			return nil
		}).
		DoOnRequest(func(n int) {
			requests++
		}).
		Subscribe(context.Background(), reactor.OnSubscribe(func(su reactor.Subscription) {
			s = su
			s.Request(1)
		}))
	<-done
	assert.Equal(t, totals, discards+nexts, "bad discards+nexts")
	assert.Equal(t, filter, nexts, "bad nexts")
	assert.Equal(t, nexts+1, requests, "bad requests")
}

func testDiscard(f flux.Flux, t *testing.T) {
	var next, next2, discard []int
	done := make(chan struct{})
	f.
		DoFinally(func(s reactor.SignalType) {
			close(done)
		}).
		DoOnNext(func(v interface{}) error {
			next = append(next, v.(int))
			return nil
		}).
		Filter(func(i interface{}) bool {
			return i.(int) > 3
		}).
		DoOnNext(func(v interface{}) error {
			next2 = append(next2, v.(int))
			return nil
		}).
		DoOnDiscard(func(i interface{}) {
			discard = append(discard, i.(int))
		}).
		Subscribe(context.Background())
	<-done
	assert.Equal(t, testData, next, "bad next")
	assert.Equal(t, len(next), len(next2)+len(discard), "bad amount")
	assert.Equal(t, []int{4, 5}, next2, "bad next2")
	assert.Equal(t, []int{1, 2, 3}, discard, "bad discard")
}

func testBlockFirst(f flux.Flux, t *testing.T) {
	var cancelled bool
	first, err := f.
		DoOnCancel(func() {
			cancelled = true
		}).
		BlockFirst(context.Background())
	assert.NoError(t, err, "block first failed")
	assert.Equal(t, testData[0], first, "bad first value")
	assert.True(t, cancelled, "should call OnCancel")
}

func testPeek(f flux.Flux, t *testing.T) {
	var complete int
	var a, b []int
	var requests int
	var ss reactor.Subscription
	done := make(chan struct{})
	f.
		DoOnNext(func(v interface{}) error {
			a = append(a, v.(int))
			return nil
		}).
		DoOnRequest(func(n int) {
			requests++
		}).
		DoOnComplete(func() {
			complete++
		}).
		DoFinally(func(s reactor.SignalType) {
			close(done)
		}).
		Subscribe(context.Background(), reactor.OnSubscribe(func(su reactor.Subscription) {
			ss = su
			ss.Request(1)
		}), reactor.OnNext(func(v interface{}) error {
			b = append(b, v.(int))
			ss.Request(1)
			return nil
		}))
	<-done
	assert.Equal(t, b, a, "values doesn't match")
	assert.Equal(t, len(a)+1, requests, "bad requests")
	assert.Equal(t, 1, complete, "bad complete")
}

func testRequest(f flux.Flux, t *testing.T) {
	var nexts []int
	var su reactor.Subscription
	done := make(chan struct{})
	f.
		DoFinally(func(s reactor.SignalType) {
			close(done)
		}).
		SubscribeOn(scheduler.Elastic()).
		Subscribe(context.Background(), reactor.OnSubscribe(func(s reactor.Subscription) {
			su = s
			su.Request(1)
		}), reactor.OnNext(func(v interface{}) error {
			nexts = append(nexts, v.(int))
			su.Request(1)
			return nil
		}))
	<-done
	assert.Equal(t, testData, nexts, "bad results")
}

func TestEmpty(t *testing.T) {
	flux.Just().Subscribe(
		context.Background(),
		reactor.OnNext(func(v interface{}) error {
			log.Println("next:", v)
			return nil
		}),
		reactor.OnComplete(func() {
			log.Println("complete")
		}),
	)
}

func TestCreateWithRequest(t *testing.T) {
	const totals = 20
	f := flux.Create(func(ctx context.Context, sink flux.Sink) {
		for i := 0; i < totals; i++ {
			sink.Next(i)
		}
		sink.Complete()
	})

	var processed int32

	su := make(chan reactor.Subscription, 1)

	sub := reactor.NewSubscriber(reactor.OnNext(func(v interface{}) error {
		log.Println("next:", v)
		processed++
		return nil
	}), reactor.OnSubscribe(func(s reactor.Subscription) {
		su <- s
		s.Request(1)
	}), reactor.OnComplete(func() {
		log.Println("complete")
	}))

	time.AfterFunc(100*time.Millisecond, func() {
		(<-su).Request(totals - 1)
	})

	f.SubscribeWith(context.Background(), sub)
	assert.Equal(t, totals, int(processed), "bad processed num")
}

func TestToChan(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
	defer cancel()
	ch, err := flux.Interval(100*time.Millisecond).ToChan(ctx, 1)
L:
	for {
		select {
		case v := <-ch:
			fmt.Println("v:", v)
		case e := <-err:
			fmt.Println("err:", e)
			break L
		}
	}
}

func TestNextWithError(t *testing.T) {
	var err error
	flux.Just(1).
		Subscribe(context.Background(), reactor.OnNext(func(v reactor.Any) error {
			return fakeErr
		}), reactor.OnError(func(e error) {
			err = e
		}))
	assert.Equal(t, fakeErr, err)
}

func TestBlockToSlice(t *testing.T) {
	results := make([]int, 0)
	err := flux.Just(3, 2, 1).BlockToSlice(context.Background(), &results)
	assert.NoError(t, err, "should not return error")
	assert.Equal(t, []int{3, 2, 1}, results)

	err = flux.Just("bad element type").BlockToSlice(context.Background(), &results)
	assert.Error(t, err, "should return error")

	err = flux.Just(1).BlockToSlice(context.Background(), nil)
	assert.Error(t, err)
}

func TestBlockToChan(t *testing.T) {
	results := make(chan int, 64)
	err := flux.Just(3, 2, 1).SubscribeOn(scheduler.Parallel()).BlockToChan(context.Background(), results)
	assert.NoError(t, err, "should not return error")
	assert.Equal(t, 3, <-results)
	assert.Equal(t, 2, <-results)
	assert.Equal(t, 1, <-results)

	err = flux.Just("bad element type").BlockToChan(context.Background(), results)
	assert.Error(t, err, "should return error")
}

func TestError(t *testing.T) {
	mockErr := errors.New("this is a mock error")
	var sig reactor.SignalType
	var e1, e2 error
	var requested int
	flux.Error(mockErr).
		DoFinally(func(s reactor.SignalType) {
			sig = s
		}).
		DoOnRequest(func(n int) {
			requested = n
		}).
		DoOnError(func(e error) {
			e1 = e
		}).
		Subscribe(
			context.Background(),
			reactor.OnNext(func(v interface{}) error {
				assert.Fail(t, "should never run here")
				return nil
			}),
			reactor.OnError(func(e error) {
				e2 = e
			}),
		)
	assert.Equal(t, mockErr, e1, "bad doOnError")
	assert.Equal(t, mockErr, e2, "bad onError")
	assert.Equal(t, reactor.SignalTypeError, sig, "bad signal")
	assert.True(t, requested > 0, "no request")
}
