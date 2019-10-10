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

func Example() {
	gen := func(ctx context.Context, sink flux.Sink) {
		for i := 0; i < 10; i++ {
			v := i
			sink.Next(v)
		}
		sink.Complete()
	}
	done := make(chan struct{})

	var su rs.Subscription
	flux.Create(gen).
		Filter(func(i interface{}) bool {
			return i.(int)%2 == 0
		}).
		Map(func(i interface{}) interface{} {
			return fmt.Sprintf("#HELLO_%04d", i.(int))
		}).
		SubscribeOn(scheduler.Elastic()).
		Subscribe(context.Background(),
			rs.OnSubscribe(func(s rs.Subscription) {
				su = s
				s.Request(1)
			}),
			rs.OnNext(func(v interface{}) {
				fmt.Println("next:", v)
				su.Request(1)
			}),
			rs.OnComplete(func() {
				close(done)
			}),
		)
	<-done
}

var testData = []int{1, 2, 3, 4, 5}

func TestSuite(t *testing.T) {
	var inputs []interface{}
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
	var su rs.Subscription
	var got int
	f.
		DoOnSubscribe(func(s rs.Subscription) {
			su = s
			su.Request(1)
		}).
		DoOnNext(func(v interface{}) {
			log.Println("next:", v)
			got++
			su.Request(1)
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
	var s rs.Subscription
	var totals, discards, nexts, requests, filter int
	done := make(chan struct{})
	f.
		DoFinally(func(s rs.SignalType) {
			assert.Equal(t, rs.SignalTypeComplete, s, "bad signal")
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
		DoOnNext(func(v interface{}) {
			nexts++
			s.Request(1)
		}).
		DoOnRequest(func(n int) {
			requests++
		}).
		Subscribe(context.Background(), rs.OnSubscribe(func(su rs.Subscription) {
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
		DoFinally(func(s rs.SignalType) {
			close(done)
		}).
		DoOnNext(func(v interface{}) {
			next = append(next, v.(int))
		}).
		Filter(func(i interface{}) bool {
			return i.(int) > 3
		}).
		DoOnNext(func(v interface{}) {
			next2 = append(next2, v.(int))
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
	var ss rs.Subscription
	done := make(chan struct{})
	f.
		DoOnNext(func(v interface{}) {
			a = append(a, v.(int))
		}).
		DoOnRequest(func(n int) {
			requests++
		}).
		DoOnComplete(func() {
			complete++
		}).
		DoFinally(func(s rs.SignalType) {
			close(done)
		}).
		Subscribe(context.Background(), rs.OnSubscribe(func(su rs.Subscription) {
			ss = su
			ss.Request(1)
		}), rs.OnNext(func(v interface{}) {
			b = append(b, v.(int))
			ss.Request(1)
		}))
	<-done
	assert.Equal(t, b, a, "values doesn't match")
	assert.Equal(t, len(a)+1, requests, "bad requests")
	assert.Equal(t, 1, complete, "bad complete")
}

func testRequest(f flux.Flux, t *testing.T) {
	var nexts []int
	var su rs.Subscription
	done := make(chan struct{})
	f.
		DoFinally(func(s rs.SignalType) {
			close(done)
		}).
		SubscribeOn(scheduler.Elastic()).
		Subscribe(context.Background(), rs.OnSubscribe(func(s rs.Subscription) {
			su = s
			su.Request(1)
		}), rs.OnNext(func(v interface{}) {
			nexts = append(nexts, v.(int))
			su.Request(1)
		}))
	<-done
	assert.Equal(t, testData, nexts, "bad results")
}

func TestEmpty(t *testing.T) {
	flux.Just().Subscribe(
		context.Background(),
		rs.OnNext(func(v interface{}) {
			log.Println("next:", v)
		}),
		rs.OnComplete(func() {
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

	su := make(chan rs.Subscription, 1)

	sub := rs.NewSubscriber(rs.OnNext(func(v interface{}) {
		log.Println("next:", v)
		processed++
	}), rs.OnSubscribe(func(s rs.Subscription) {
		su <- s
		s.Request(1)
	}), rs.OnComplete(func() {
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

func TestError(t *testing.T) {
	mockErr := errors.New("this is a mock error")
	var sig rs.SignalType
	var e1, e2 error
	var requested int
	flux.Error(mockErr).
		DoFinally(func(s rs.SignalType) {
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
			rs.OnNext(func(v interface{}) {
				assert.Fail(t, "should never run here")
			}),
			rs.OnError(func(e error) {
				e2 = e
			}),
		)
	assert.Equal(t, mockErr, e1, "bad doOnError")
	assert.Equal(t, mockErr, e2, "bad onError")
	assert.Equal(t, rs.SignalTypeError, sig, "bad signal")
	assert.True(t, requested > 0, "no request")
}
