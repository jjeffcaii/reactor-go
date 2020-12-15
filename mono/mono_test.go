package mono_test

import (
	"context"
	"fmt"
	"sync/atomic"
	"testing"
	"time"

	"github.com/jjeffcaii/reactor-go"
	"github.com/jjeffcaii/reactor-go/hooks"
	"github.com/jjeffcaii/reactor-go/mono"
	"github.com/jjeffcaii/reactor-go/scheduler"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

type Any = reactor.Any

func Example() {
	gen := func(ctx context.Context, sink mono.Sink) {
		sink.Success("World")
	}
	mono.
		Create(gen).
		Map(func(i Any) (o Any, err error) {
			o = "Hello " + i.(string) + "!"
			return
		}).
		DoOnNext(func(v Any) error {
			fmt.Println(v)
			return nil
		}).
		Subscribe(context.Background())
}

// Should print
// Hello World!

const num = 333

var fakeErr = errors.New("fake error")

func TestCreate_Panic(t *testing.T) {
	var err error
	mono.
		Create(func(i context.Context, sink mono.Sink) {
			panic("mock error")
		}).
		DoOnError(func(e error) {
			err = e
		}).
		SubscribeOn(scheduler.Immediate()).
		Subscribe(context.Background())
	assert.Error(t, err)
}

func TestDelay(t *testing.T) {
	begin := time.Now()
	v, err := mono.Delay(10 * time.Millisecond).Block(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, 1, int(time.Since(begin).Nanoseconds()/1e7))
	assert.Equal(t, int64(0), v)

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	_, err = mono.Delay(300 * time.Millisecond).Block(ctx)
	assert.Error(t, err)
	cancel()
}

func TestSuite(t *testing.T) {
	// TODO: processor
	//pc := mono.CreateProcessor()
	//go func() {
	//	pc.Success(num)
	//}()
	all := map[string]mono.Mono{
		"Just": mono.Just(num),
		"Create": mono.Create(func(i context.Context, sink mono.Sink) {
			sink.Success(num)
		}),
		//"Processor": pc,
	}

	for k, v := range all {
		t.Run(k+"_Map", func(t *testing.T) {
			testMap(v, t)
		})
		t.Run(k+"_Filter", func(t *testing.T) {
			testFilter(v, t)
		})
		t.Run(k+"_DelayElement", func(t *testing.T) {
			testDelayElement(v, t)
		})
		t.Run(k+"_Panic", func(t *testing.T) {
			testPanic(v, t)
		})
		t.Run(k+"_SubscribeOn", func(t *testing.T) {
			testSubscribeOn(v, t)
		})
		t.Run(k+"_FlatMap", func(t *testing.T) {
			testFlatMap(v, t)
		})
		t.Run(k+"_Cancel", func(t *testing.T) {
			testCancel(v, t)
		})
		t.Run(k+"_ContextDone", func(t *testing.T) {
			testContextDone(v, t)
		})
		t.Run(k+"_DoOnSubscribe", func(t *testing.T) {
			testDoOnSubscribe(v, t)
		})
	}
}

func testDoOnSubscribe(m mono.Mono, t *testing.T) {
	var called bool
	v, err := m.DoOnSubscribe(func(ctx context.Context, su reactor.Subscription) {
		called = true
	}).Block(context.Background())
	assert.NoError(t, err, "oops")
	assert.Equal(t, num, v, "bad value")
	assert.True(t, called, "doOnSubscribe hasn't been called")
}

func testPanic(m mono.Mono, t *testing.T) {
	fakeErr := errors.New("mock next panic")
	checker := func(in mono.Mono) {
		var catches error
		in.DoOnError(func(e error) {
			catches = e
		}).Subscribe(context.Background())
		assert.Equal(t, fakeErr, errors.Cause(catches), "not that error")
	}
	checker(m.DoOnNext(func(v Any) error {
		return fakeErr
	}))
	checker(m.Map(func(i Any) (Any, error) {
		return nil, fakeErr
	}))
	checker(m.Filter(func(i Any) bool {
		panic(fakeErr)
	}))
}

func testDelayElement(m mono.Mono, t *testing.T) {
	var begin time.Time
	m.DelayElement(1*time.Second).
		DoOnNext(func(v Any) error {
			assert.Equal(t, num, v, "bad next value")
			assert.Equal(t, 1, int(time.Since(begin).Seconds()), "bad passed time")
			return nil
		}).
		Subscribe(context.Background(), reactor.OnSubscribe(func(ctx context.Context, su reactor.Subscription) {
			begin = time.Now()
			su.Request(reactor.RequestInfinite)
		}))
}

func testMap(m mono.Mono, t *testing.T) {
	v, err := m.
		DoOnNext(func(v Any) error {
			assert.Equal(t, num, v, "bad value before map")
			return nil
		}).
		Map(func(i Any) (Any, error) {
			return i.(int) * 2, nil
		}).
		DoOnNext(func(v Any) error {
			assert.Equal(t, num*2, v, "bad value after map")
			return nil
		}).
		Block(context.Background())
	assert.NoError(t, err, "err occurred")
	assert.Equal(t, num*2, v.(int), "bad map result")

	_, err = m.
		Map(func(any reactor.Any) (reactor.Any, error) {
			return nil, fakeErr
		}).
		Block(context.Background())
	assert.Equal(t, fakeErr, err, "should return error")
}

func testFilter(m mono.Mono, t *testing.T) {
	done := make(chan struct{})
	var discard bool
	var next bool
	var step []string
	m.
		DoFinally(func(signal reactor.SignalType) {
			step = append(step, "finally")
			assert.Equal(t, reactor.SignalTypeComplete, signal)
			close(done)
		}).
		DoOnNext(func(v Any) error {
			step = append(step, "next")
			next = true
			assert.Equal(t, num, v, "bad next value")
			return nil
		}).
		Filter(func(i Any) bool {
			step = append(step, "filter")
			return i.(int) > num
		}).
		DoOnDiscard(func(i Any) {
			step = append(step, "discard")
			discard = true
			assert.Equal(t, num, i, "bad discard value")
		}).
		Subscribe(context.Background())
	assert.True(t, next, "missing next")
	assert.True(t, discard, "missing discard")
	assert.Equal(t, "finally", step[len(step)-1], "bad finally order")
}

func testSubscribeOn(m mono.Mono, t *testing.T) {
	done := make(chan struct{})
	m.
		SubscribeOn(scheduler.Elastic()).
		DoOnNext(func(v Any) error {
			assert.Equal(t, num, v, "bad next value")
			return nil
		}).
		DoOnComplete(func() {
			close(done)
		}).
		Subscribe(context.Background())
	<-done
}

func testFlatMap(m mono.Mono, t *testing.T) {
	v, err := m.
		FlatMap(func(i Any) mono.Mono {
			return mono.Just(i).
				Map(func(i Any) (Any, error) {
					time.Sleep(200 * time.Millisecond)
					return i.(int) * 2, nil
				}).
				SubscribeOn(scheduler.Elastic())
		}).
		Block(context.Background())
	assert.NoError(t, err, "an error occurred")
	assert.Equal(t, num*2, v, "bad result")
}

func testCancel(m mono.Mono, t *testing.T) {
	var cancelled bool
	m.
		DoOnNext(func(v Any) error {
			assert.Fail(t, "should never run here")
			return nil
		}).
		DoOnCancel(func() {
			cancelled = true
		}).
		Subscribe(context.Background(), reactor.OnSubscribe(func(ctx context.Context, su reactor.Subscription) {
			su.Cancel()
		}))
	assert.True(t, cancelled, "bad cancelled")
}

func testContextDone(m mono.Mono, t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	var err error
	m.
		DoOnNext(func(v Any) error {
			assert.Fail(t, "should never run here")
			return nil
		}).
		DoOnError(func(e error) {
			err = e
		}).
		Subscribe(ctx)
	assert.Error(t, err, "should catch error")
}

func TestError(t *testing.T) {
	mockErr := errors.New("this is a mock error")
	var sig reactor.SignalType
	var e1, e2 error
	mono.Error(mockErr).
		DoFinally(func(s reactor.SignalType) {
			sig = s
		}).
		DoOnError(func(e error) {
			e1 = e
		}).
		Subscribe(
			context.Background(),
			reactor.OnNext(func(v Any) error {
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
}

func TestJustOrEmpty(t *testing.T) {
	v, err := mono.JustOrEmpty(nil).Block(context.Background())
	assert.NoError(t, err, "should not return error")
	assert.Nil(t, v, "should return nil")
	v, err = mono.JustOrEmpty(1).Block(context.Background())
	assert.NoError(t, err, "should not return error")
	assert.Equal(t, 1, v, "value doesn't match")
}

func TestMapWithError(t *testing.T) {
	_, err := mono.Error(fakeErr).
		Map(func(any reactor.Any) (reactor.Any, error) {
			assert.Fail(t, "should never run here")
			return 1234, nil
		}).
		Block(context.Background())
	assert.Equal(t, fakeErr, err, "should return error")
}

func TestTimeout(t *testing.T) {
	fakeErr := errors.New("fake err")
	dropped := new(int32)
	errorDropped := new(int32)
	hooks.OnNextDrop(func(v reactor.Any) {
		atomic.AddInt32(dropped, 1)
	})
	hooks.OnErrorDrop(func(e error) {
		atomic.AddInt32(errorDropped, 1)
	})
	_, err := mono.
		Create(func(ctx context.Context, s mono.Sink) {
			time.Sleep(200 * time.Millisecond)
			s.Success("hello")
		}).
		Timeout(100 * time.Millisecond).
		Block(context.Background())
	assert.Error(t, err)
	time.Sleep(200 * time.Millisecond)
	assert.Equal(t, int32(1), atomic.LoadInt32(dropped))

	value, err := mono.Just("hello").Timeout(100 * time.Millisecond).Block(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, "hello", value)

	_, err = mono.Error(fakeErr).Timeout(100 * time.Millisecond).Block(context.Background())
	assert.Equal(t, fakeErr, err)

	_, err = mono.Create(func(ctx context.Context, s mono.Sink) {
		time.Sleep(100 * time.Millisecond)
		s.Error(err)
	}).Timeout(500 * time.Millisecond).Block(context.Background())
	assert.Equal(t, fakeErr, err)

	_, err = mono.Create(func(ctx context.Context, s mono.Sink) {
		time.Sleep(100 * time.Millisecond)
		s.Error(err)
	}).Timeout(50 * time.Millisecond).Block(context.Background())
	assert.True(t, reactor.IsCancelledError(err))
	assert.Equal(t, int32(1), atomic.LoadInt32(errorDropped))
}

func TestBlock(t *testing.T) {
	v, err := mono.
		Create(func(ctx context.Context, s mono.Sink) {
			s.Success(1)
		}).
		SubscribeOn(scheduler.Parallel()).
		Block(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, 1, v)

	v, err = mono.Create(func(ctx context.Context, s mono.Sink) {
		s.Success(1)
	}).Filter(func(any reactor.Any) bool {
		return false
	}).Block(context.Background())
	assert.NoError(t, err)
	assert.Nil(t, v)
}

func TestOneshot(t *testing.T) {
	for _, m := range []mono.Mono{
		mono.JustOneshot(1),
		mono.CreateOneshot(func(ctx context.Context, s mono.Sink) {
			s.Success(1)
		}),
	} {
		result, err := m.
			Map(func(any reactor.Any) (reactor.Any, error) {
				return any.(int) * 2, nil
			}).
			DoOnNext(func(v reactor.Any) error {
				assert.Equal(t, 2, v.(int))
				return nil
			}).
			DoOnError(func(e error) {
				assert.FailNow(t, "unreachable")
			}).
			Block(context.Background())
		assert.NoError(t, err)
		assert.Equal(t, 2, result)
	}

	n, err := mono.JustOneshot(1).
		FlatMap(func(any reactor.Any) mono.Mono {
			return mono.Just(any.(int) * 3)
		}).
		Block(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, 3, n)

	discardCalls := new(int32)

	mono.
		CreateOneshot(func(ctx context.Context, s mono.Sink) {
			s.Success(111)
		}).
		Filter(func(any reactor.Any) bool {
			return any.(int) > 222
		}).
		DoOnDiscard(func(v reactor.Any) {
			assert.Equal(t, 111, v)
			atomic.AddInt32(discardCalls, 1)
		}).
		SwitchIfEmpty(mono.Just(333)).
		DoOnNext(func(v reactor.Any) error {
			assert.Equal(t, 333, v)
			return nil
		}).
		Subscribe(context.Background())

	assert.Equal(t, int32(1), atomic.LoadInt32(discardCalls))

	_, err = mono.
		CreateOneshot(func(ctx context.Context, s mono.Sink) {
			time.Sleep(100 * time.Millisecond)
			s.Success(1)
		}).
		Timeout(50 * time.Millisecond).
		Block(context.Background())
	assert.Error(t, err)
	assert.True(t, reactor.IsCancelledError(err))

	done := make(chan struct{})
	now := time.Now()
	mono.Just(1).
		DoFinally(func(s reactor.SignalType) {
			close(done)
		}).
		DelayElement(100*time.Millisecond).
		Subscribe(context.Background(), reactor.OnNext(func(v reactor.Any) error {
			assert.True(t, time.Since(now)/1e6 >= 100)
			return nil
		}))
	<-done
}

func TestErrorOneshot(t *testing.T) {
	fakeErr := errors.New("fake error")
	finallyCalls := new(int32)
	subscribeCalls := new(int32)
	_, err := mono.
		ErrorOneshot(fakeErr).
		DoFinally(func(s reactor.SignalType) {
			atomic.AddInt32(finallyCalls, 1)
		}).
		DoOnSubscribe(func(ctx context.Context, su reactor.Subscription) {
			atomic.AddInt32(subscribeCalls, 1)
		}).
		DoOnCancel(func() {
			assert.FailNow(t, "unreachable")
		}).
		DoOnNext(func(v reactor.Any) error {
			assert.FailNow(t, "unreachable")
			return nil
		}).
		DoOnError(func(e error) {
			assert.Equal(t, fakeErr, e)
		}).
		SubscribeOn(scheduler.Parallel()).
		Block(context.Background())
	assert.Error(t, err, "should return error")
	assert.Equal(t, fakeErr, err)
	assert.Equal(t, int32(1), atomic.LoadInt32(finallyCalls))
	assert.Equal(t, int32(1), atomic.LoadInt32(subscribeCalls))
}

func TestIsSubscribeOnParallel(t *testing.T) {
	assert.False(t, mono.IsSubscribeAsync(mono.Just(1)))
	assert.True(t, mono.IsSubscribeAsync(mono.Just(1).SubscribeOn(scheduler.Parallel())))
	assert.True(t, mono.IsSubscribeAsync(mono.Just(1).SubscribeOn(scheduler.Single())))
	assert.True(t, mono.IsSubscribeAsync(mono.JustOneshot(1).SubscribeOn(scheduler.Elastic())))
}

func TestJust(t *testing.T) {
	assert.Panics(t, func() {
		mono.Just(nil)
	})
	assert.Panics(t, func() {
		mono.JustOneshot(nil)
	})
	assert.NotNil(t, mono.Just(1))
	assert.NotNil(t, mono.JustOneshot(1))
}

func TestDefaultIfEmpty(t *testing.T) {
	value, err := mono.Empty().DefaultIfEmpty(333).Block(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, 333, value, "bad result")
}
