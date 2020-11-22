package mono_test

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/jjeffcaii/reactor-go"
	"github.com/jjeffcaii/reactor-go/mono"
	"github.com/jjeffcaii/reactor-go/scheduler"
	"github.com/stretchr/testify/assert"
)

func TestProcessor(t *testing.T) {
	p, s, d := mono.NewProcessor(nil, nil)
	defer d.Dispose()

	go func() {
		now := time.Now()
		s.Success("hello world")
		t.Log("cost:", time.Since(now))
	}()

	const total = 8

	var wg sync.WaitGroup

	wg.Add(total)

	onNext := reactor.OnNext(func(v reactor.Any) error {
		time.Sleep(100 * time.Millisecond)
		t.Log("next:", v)
		return nil
	})

	onComplete := reactor.OnComplete(func() {
		wg.Done()
	})

	for range [total]struct{}{} {
		p.SubscribeOn(scheduler.Parallel()).Subscribe(context.Background(), onNext, onComplete)
	}

	wg.Wait()

}
func TestProcessor_Map(t *testing.T) {
	p, s, d := mono.NewProcessor(nil, nil)
	defer d.Dispose()

	time.AfterFunc(100*time.Millisecond, func() {
		s.Success(1024)
	})

	transform := func(any reactor.Any) (reactor.Any, error) {
		return any.(int) * 2, nil
	}
	v, err := p.Map(transform).Block(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, 2048, v)
}

func TestProcessorNG_Order(t *testing.T) {
	valueChan := make(chan reactor.Any)
	onNext := reactor.OnNext(func(v reactor.Any) error {
		valueChan <- v
		return nil
	})

	p, s, d := mono.NewProcessor(nil, nil)
	defer d.Dispose()
	p.SubscribeOn(scheduler.Parallel()).Subscribe(context.Background(), onNext)

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
	p, _, d := mono.NewProcessor(nil, nil)
	defer d.Dispose()
	onSubscribe := reactor.OnSubscribe(func(ctx context.Context, su reactor.Subscription) {
		su.Cancel()
	})
	doFinally := func(s reactor.SignalType) {
		close(done)
	}
	onNext := reactor.OnNext(func(v reactor.Any) error {
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
		p, s, _ := mono.NewProcessor(scheduler.Parallel(), func(signalType reactor.SignalType, disposable reactor.Disposable) {
			disposable.Dispose()
			wg.Done()
		})

		go func() {
			s.Success(fakeValue)
		}()

		p.Subscribe(context.Background(), reactor.OnNext(func(v reactor.Any) error {
			assert.Equal(t, fakeValue, v)
			return nil
		}), reactor.OnError(func(e error) {
			assert.NoError(t, e, "should not return error")
		}), reactor.OnComplete(func() {
			atomic.AddInt32(cnt, 1)
		}))
	}
	wg.Wait()

	assert.Equal(t, int32(total), atomic.LoadInt32(cnt))
}
