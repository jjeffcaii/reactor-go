package flux_test

import (
	"context"
	"errors"
	"fmt"
	"sync/atomic"
	"testing"
	"time"

	"github.com/jjeffcaii/reactor-go"
	"github.com/jjeffcaii/reactor-go/flux"
	"github.com/jjeffcaii/reactor-go/scheduler"
	"github.com/stretchr/testify/assert"
)

type Any = reactor.Any

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

var fakeErr = errors.New("fake error")
var testData = []Any{1, 2, 3, 4, 5}

func init() {
	flux.InitBuffSize(flux.BuffSizeXS)
}

func TestSuite(t *testing.T) {
	all := make(map[string]func() flux.Flux)
	var instJust = flux.Just(testData...)
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
		t.Run(fmt.Sprintf("%s_DoSubscribeOn", k), func(t *testing.T) {
			testDoOnSubscribe(gen(), t)
		})
	}
}

func testDoOnSubscribe(f flux.Flux, t *testing.T) {
	var su reactor.Subscription
	var got int32
	onNext := func(v Any) error {
		atomic.AddInt32(&got, 1)
		su.Request(1)
		return nil
	}
	onSubscribe := func(s reactor.Subscription) {
		su = s
		su.Request(1)
	}
	_, err := f.DoOnNext(onNext).DoOnSubscribe(onSubscribe).BlockLast(context.Background())
	assert.NoError(t, err)
	assert.Len(t, testData, int(atomic.LoadInt32(&got)), "bad len")
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
	var next, next2, discard []Any
	done := make(chan struct{})
	f.
		DoFinally(func(s reactor.SignalType) {
			close(done)
		}).
		DoOnNext(func(v Any) error {
			next = append(next, v)
			return nil
		}).
		Filter(func(i Any) bool {
			return i.(int) > 3
		}).
		DoOnNext(func(v Any) error {
			next2 = append(next2, v)
			return nil
		}).
		DoOnDiscard(func(i Any) {
			discard = append(discard, i)
		}).
		Subscribe(context.Background())
	<-done
	assert.Equal(t, testData, next, "bad next")
	assert.Equal(t, len(next), len(next2)+len(discard), "bad amount")
	assert.Equal(t, []Any{4, 5}, next2, "bad next2")
	assert.Equal(t, []Any{1, 2, 3}, discard, "bad discard")
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
	var nexts []Any
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
		}), reactor.OnNext(func(v Any) error {
			nexts = append(nexts, v)
			su.Request(1)
			return nil
		}))
	<-done
	assert.Equal(t, testData, nexts, "bad results")
}

func TestEmpty(t *testing.T) {
	innerTestEmpty(t, flux.Just())
	innerTestEmpty(t, flux.Empty())
}

func innerTestEmpty(t *testing.T, empty flux.Flux) {
	completed := int32(0)
	empty.Subscribe(
		context.Background(),
		reactor.OnNext(func(v Any) error {
			assert.Fail(t, "should be unreachable")
			return nil
		}),
		reactor.OnComplete(func() {
			atomic.AddInt32(&completed, 1)
		}),
	)
	assert.Equal(t, int32(1), atomic.LoadInt32(&completed))
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
		fmt.Println("next:", v)
		processed++
		return nil
	}), reactor.OnSubscribe(func(s reactor.Subscription) {
		su <- s
		s.Request(1)
	}), reactor.OnComplete(func() {
		fmt.Println("complete")
	}))

	time.AfterFunc(100*time.Millisecond, func() {
		(<-su).Request(totals - 1)
	})

	f.SubscribeWith(context.Background(), sub)
	assert.Equal(t, totals, int(processed), "bad processed num")
}

func TestNextWithError(t *testing.T) {
	var err error
	flux.Just(1).
		Subscribe(context.Background(), reactor.OnNext(func(v Any) error {
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

func TestSubscribeWithChan(t *testing.T) {
	valChan := make(chan int)
	errChan := make(chan error)
	flux.Just(testData...).
		DoFinally(func(s reactor.SignalType) {
			close(valChan)
			close(errChan)
		}).
		SubscribeOn(scheduler.Parallel()).
		SubscribeWithChan(context.Background(), valChan, errChan)

	var actualData []int
L:
	for {
		select {
		case value, ok := <-valChan:
			if !ok {
				break L
			}
			actualData = append(actualData, value)
		case _, ok := <-errChan:
			if !ok {
				break L
			}
			assert.Fail(t, "should never run here")
		}
	}
	assert.Len(t, actualData, len(testData))
	for i := 0; i < len(testData); i++ {
		assert.Equal(t, testData[i], actualData[i], "result doesn't match")
	}

}

func TestSubscribeWithChan_Broken(t *testing.T) {
	valChan := make(chan int, 1)
	errChan := make(chan error, 1)
	flux.Just("bad element type").
		SubscribeOn(scheduler.Parallel()).
		SubscribeWithChan(context.Background(), valChan, errChan)
	select {
	case <-valChan:
		assert.Fail(t, "should be unreachable")
	case err := <-errChan:
		assert.Error(t, err, "should return error")
	}
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

func TestMapFromErrorFlux(t *testing.T) {
	_, err := flux.Error(fakeErr).
		Map(func(any Any) (Any, error) {
			assert.Fail(t, "should be unreachable")
			return 1234, nil
		}).
		BlockLast(context.Background())
	assert.Equal(t, fakeErr, err, "should return error")
}

func TestMapWithError(t *testing.T) {
	_, err := flux.Just(134).
		Map(func(_ Any) (Any, error) {
			return nil, fakeErr
		}).
		BlockLast(context.Background())
	assert.Equal(t, fakeErr, err, "should return error")
}

func TestDoOnNextWithError(t *testing.T) {
	var e1 error
	_, err := flux.Just(1).
		DoOnNext(func(v Any) error {
			return fakeErr
		}).
		DoOnError(func(e error) {
			e1 = e
		}).
		BlockLast(context.Background())
	assert.Equal(t, fakeErr, e1)
	assert.Equal(t, fakeErr, err)
}

func TestDelayElement(t *testing.T) {
	start := time.Now()
	_, err := flux.Just(1, 2, 3).DelayElement(100 * time.Millisecond).BlockLast(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, 3, int(time.Since(start).Milliseconds()/100))

	start = time.Now()
	_, err = flux.Error(fakeErr).DelayElement(100 * time.Millisecond).BlockLast(context.Background())
	assert.Equal(t, fakeErr, err)
	assert.True(t, time.Since(start) < 100*time.Millisecond)
}
