package mono_test

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/jjeffcaii/reactor-go"
	mono "github.com/jjeffcaii/reactor-go/mono2"
	"github.com/jjeffcaii/reactor-go/scheduler"
)

func TestProcessor(t *testing.T) {
	p, s, d := mono.NewProcessor[string]()
	defer d.Dispose()

	go func() {
		now := time.Now()
		s.Success("hello world")
		t.Log("cost:", time.Since(now))
	}()

	const total = 8

	var wg sync.WaitGroup

	wg.Add(total)

	onNext := reactor.OnNext[string](func(ctx context.Context, item string) error {
		time.Sleep(100 * time.Millisecond)
		t.Log("onNext:", item)
		return nil
	})

	onComplete := reactor.OnComplete[string](func(ctx context.Context) {
		wg.Done()
	})

	for range [total]struct{}{} {
		p.SubscribeOn(scheduler.Parallel()).Subscribe(context.Background(), onNext, onComplete)
	}

	wg.Wait()

}
func TestProcessor_Map(t *testing.T) {
	p, s, d := mono.NewProcessor[int]()
	defer d.Dispose()

	time.AfterFunc(100*time.Millisecond, func() {
		s.Success(1024)
	})

	mul2 := func(ctx context.Context, next int) (int, error) {
		return 2 * next, nil
	}
	v, err := p.Map(mul2).Block(context.TODO())
	assert.NoError(t, err)
	assert.True(t, v.Valid)
	assert.Equal(t, 2048, v.V)
}

func TestProcessorNG_Order(t *testing.T) {
	valueChan := make(chan int)

	p, s, d := mono.NewProcessor[int](nil, nil)
	defer d.Dispose()
	p.SubscribeOn(scheduler.Parallel()).
		Subscribe(
			context.Background(),
			reactor.OnNext[int](func(ctx context.Context, v int) error {
				valueChan <- v
				return nil
			}),
		)

	select {
	case <-valueChan:
		assert.Fail(t, "unreachable")
	default:
	}

	s.Success(123)
	assert.Equal(t, 123, <-valueChan)
}

func TestProcessorSubscriberNG_Cancel(t *testing.T) {
	done := make(chan struct{})
	p, _, d := mono.NewProcessor[int](nil, nil)
	defer d.Dispose()
	onSubscribe := reactor.OnSubscribe[int](func(ctx context.Context, su reactor.Subscription) {
		su.Cancel(ctx)
	})
	doFinally := func(ctx context.Context, s reactor.SignalType) {
		close(done)
	}
	onNext := reactor.OnNext[int](func(ctx context.Context, v int) error {
		assert.Fail(t, "unreachable")
		return nil
	})
	p.DoFinally(doFinally).Subscribe(context.Background(), onSubscribe, onNext)
}

func TestProcessor_Finally(t *testing.T) {
	const total = 2

	var wg sync.WaitGroup
	wg.Add(total)

	cnt := new(int32)
	fakeValue := 1

	for range [total]struct{}{} {
		p, s, _ := mono.NewProcessor[int](
			mono.WithProcessorScheduler[int](scheduler.Parallel()),
			mono.WithProcessorFinallyHook[int](func(signalType reactor.SignalType, disposable reactor.Disposable) {
				disposable.Dispose()
				wg.Done()
			}),
		)

		go func() {
			s.Success(fakeValue)
		}()

		p.Subscribe(context.Background(), reactor.OnNext[int](func(ctx context.Context, v int) error {
			assert.Equal(t, fakeValue, v)
			return nil
		}), reactor.OnError[int](func(ctx context.Context, e error) {
			assert.NoError(t, e, "should not return error")
		}), reactor.OnComplete[int](func(ctx context.Context) {
			atomic.AddInt32(cnt, 1)
		}))
	}
	wg.Wait()

	assert.Equal(t, int32(total), atomic.LoadInt32(cnt))
}
