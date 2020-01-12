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

func Example() {
	gen := func(ctx context.Context, sink mono.Sink) {
		sink.Success("World")
	}
	mono.
		Create(gen).
		Map(func(i interface{}) interface{} {
			return "Hello " + i.(string) + "!"
		}).
		DoOnNext(func(v interface{}) {
			fmt.Println(v)
		}).
		Subscribe(context.Background())
}

// Should print
// Hello World!

const num = 333

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
}

func TestSwitchIfEmpty(t *testing.T) {
	v, err := mono.Just(num).SwitchIfEmpty(mono.Just(num * 2)).Block(context.Background())
	assert.NoError(t, err, "err occurred")
	assert.Equal(t, num, v, "bad result")
	v, err = mono.Empty().
		SwitchIfEmpty(mono.Just(num)).
		Map(func(i interface{}) interface{} {
			return i.(int) * 2
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
	v, err := m.DoOnSubscribe(func(su rs.Subscription) {
		called = true
	}).Block(context.Background())
	assert.NoError(t, err, "oops")
	assert.Equal(t, num, v, "bad value")
	assert.True(t, called, "doOnSubscribe hasn't been called")
}

func testPanic(m mono.Mono, t *testing.T) {
	mock := errors.New("mock next panic")
	checker := func(in mono.Mono) {
		var catches error
		in.DoOnError(func(e error) {
			catches = e
		}).Subscribe(context.Background())
		assert.Equal(t, mock, catches, "not that error")
	}
	checker(m.DoOnNext(func(v interface{}) {
		panic(mock)
	}))
	checker(m.Map(func(i interface{}) interface{} {
		panic(mock)
	}))
	checker(m.Filter(func(i interface{}) bool {
		panic(mock)
	}))
}

func testDelayElement(m mono.Mono, t *testing.T) {
	var begin time.Time
	m.DelayElement(1*time.Second).
		DoOnNext(func(v interface{}) {
			assert.Equal(t, num, v, "bad next value")
			assert.Equal(t, 1, int(time.Since(begin).Seconds()), "bad passed time")
		}).
		Subscribe(context.Background(), rs.OnSubscribe(func(su rs.Subscription) {
			begin = time.Now()
			su.Request(rs.RequestInfinite)
		}))
}

func testMap(m mono.Mono, t *testing.T) {
	v, err := m.
		DoOnNext(func(v interface{}) {
			assert.Equal(t, num, v, "bad value before map")
		}).
		Map(func(i interface{}) interface{} {
			return i.(int) * 2
		}).
		DoOnNext(func(v interface{}) {
			assert.Equal(t, num*2, v, "bad value after map")
		}).
		Block(context.Background())
	assert.NoError(t, err, "err occurred")
	assert.Equal(t, num*2, v.(int), "bad map result")
}

func testFilter(m mono.Mono, t *testing.T) {
	done := make(chan struct{})
	var discard bool
	var next bool
	var step []string
	m.
		DoFinally(func(signal rs.SignalType) {
			step = append(step, "finally")
			assert.Equal(t, rs.SignalTypeComplete, signal)
			close(done)
		}).
		DoOnNext(func(v interface{}) {
			step = append(step, "next")
			next = true
			assert.Equal(t, num, v, "bad next value")
		}).
		Filter(func(i interface{}) bool {
			step = append(step, "filter")
			return i.(int) > num
		}).
		DoOnDiscard(func(i interface{}) {
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
		DoOnNext(func(v interface{}) {
			assert.Equal(t, num, v, "bad next value")
		}).
		DoOnComplete(func() {
			close(done)
		}).
		Subscribe(context.Background())
	<-done
}

func testFlatMap(m mono.Mono, t *testing.T) {
	v, err := m.
		FlatMap(func(i interface{}) mono.Mono {
			return mono.Just(i).
				Map(func(i interface{}) interface{} {
					time.Sleep(200 * time.Millisecond)
					return i.(int) * 2
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
		DoOnNext(func(v interface{}) {
			assert.Fail(t, "should never run here")
		}).
		DoOnCancel(func() {
			cancelled = true
		}).
		Subscribe(context.Background(), rs.OnSubscribe(func(su rs.Subscription) {
			su.Cancel()
		}))
	assert.True(t, cancelled, "bad cancelled")
}

func testContextDone(m mono.Mono, t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	var catches error
	m.
		DoOnNext(func(v interface{}) {
			assert.Fail(t, "should never run here")
		}).
		DoOnError(func(e error) {
			catches = e
		}).
		Subscribe(ctx)
	assert.Equal(t, rs.ErrSubscribeCancelled, catches, "bad cancelled error")
}

func BenchmarkNative(b *testing.B) {
	var sum int64
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			var v interface{} = int64(1)
			atomic.AddInt64(&sum, v.(int64))
		}
	})
}

func BenchmarkJust(b *testing.B) {
	var sum int64
	m := mono.Just(int64(1))
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		s := rs.NewSubscriber(rs.OnNext(func(v interface{}) {
			atomic.AddInt64(&sum, v.(int64))
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
		s := rs.NewSubscriber(rs.OnNext(func(v interface{}) {
			atomic.AddInt64(&sum, v.(int64))
		}))
		for pb.Next() {
			m.SubscribeWith(context.Background(), s)
		}
	})
}

func TestError(t *testing.T) {
	mockErr := errors.New("this is a mock error")
	var sig rs.SignalType
	var e1, e2 error
	mono.Error(mockErr).
		DoFinally(func(s rs.SignalType) {
			sig = s
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
}
