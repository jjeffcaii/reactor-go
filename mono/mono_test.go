package mono_test

import (
	"context"
	"errors"
	"fmt"
	"sync/atomic"
	"testing"
	"time"

	"github.com/jjeffcaii/reactor-go"
	"github.com/jjeffcaii/reactor-go/mono"
	"github.com/jjeffcaii/reactor-go/scheduler"
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
	mock := errors.New("mock error")
	var catches error
	mono.
		Create(func(i context.Context, sink mono.Sink) {
			panic(mock)
		}).
		DoOnError(func(e error) {
			catches = e
		}).
		SubscribeOn(scheduler.Immediate()).
		Subscribe(context.Background())
	assert.Equal(t, mock, catches, "bad catches")
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

func TestSwitchIfEmpty(t *testing.T) {
	v, err := mono.Just(num).SwitchIfEmpty(mono.Just(num * 2)).Block(context.Background())
	assert.NoError(t, err, "err occurred")
	assert.Equal(t, num, v, "bad result")
	v, err = mono.Empty().
		SwitchIfEmpty(mono.Just(num)).
		Map(func(i Any) (Any, error) {
			return i.(int) * 2, nil
		}).
		Block(context.Background())
	assert.NoError(t, err, "err occurred")
	assert.Equal(t, 666, v.(int), "bad result")
}

func TestSuite(t *testing.T) {
	pc := mono.CreateProcessor()
	all := map[string]mono.Mono{
		"Just": mono.Just(num),
		"Create": mono.Create(func(i context.Context, sink mono.Sink) {
			sink.Success(num)
		}),
		"Processor": pc,
	}

	go func() {
		pc.Success(num)
	}()

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
		assert.Equal(t, fakeErr, catches, "not that error")
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
	var catches error
	m.
		DoOnNext(func(v Any) error {
			assert.Fail(t, "should never run here")
			return nil
		}).
		DoOnError(func(e error) {
			catches = e
		}).
		Subscribe(ctx)
	assert.Equal(t, reactor.ErrSubscribeCancelled, catches, "bad cancelled error")
}

func BenchmarkNative(b *testing.B) {
	var sum int64
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			var v Any = int64(1)
			atomic.AddInt64(&sum, v.(int64))
		}
	})
}

func BenchmarkJust(b *testing.B) {
	var sum int64
	m := mono.Just(int64(1))
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		s := reactor.NewSubscriber(reactor.OnNext(func(v Any) error {
			atomic.AddInt64(&sum, v.(int64))
			return nil
		}))
		for pb.Next() {
			m.SubscribeWith(context.Background(), s)
		}
	})
}

func BenchmarkCreate(b *testing.B) {
	var sum int64
	m := mono.Create(func(i context.Context, sink mono.Sink) {
		sink.Success(int64(1))
	})
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		s := reactor.NewSubscriber(reactor.OnNext(func(v Any) error {
			atomic.AddInt64(&sum, v.(int64))
			return nil
		}))
		for pb.Next() {
			m.SubscribeWith(context.Background(), s)
		}
	})
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
