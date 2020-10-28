package mono_test

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/jjeffcaii/reactor-go"
	"github.com/jjeffcaii/reactor-go/hooks"
	"github.com/jjeffcaii/reactor-go/mono"
	"github.com/jjeffcaii/reactor-go/scheduler"
	"github.com/stretchr/testify/assert"
)

func TestProcessor(t *testing.T) {
	p := mono.CreateProcessor()

	time.AfterFunc(100*time.Millisecond, func() {
		p.Success(333)
	})

	v, err := p.
		Map(func(i Any) (Any, error) {
			return i.(int) * 2, nil
		}).
		Block(context.Background())
	assert.NoError(t, err, "block failed")
	assert.Equal(t, 666, v.(int), "bad result")

	assert.Panics(t, func() {
		p.Subscribe(context.Background())
	})
}

func TestProcessorOneshot(t *testing.T) {
	m, s := mono.CreateProcessorOneshot()
	time.AfterFunc(100*time.Millisecond, func() {
		s.Success(333)
	})

	v, err := m.
		Map(func(i Any) (Any, error) {
			return i.(int) * 2, nil
		}).
		Block(context.Background())
	assert.NoError(t, err, "block failed")
	assert.Equal(t, 666, v.(int), "bad result")

	assert.Panics(t, func() {
		m.Subscribe(context.Background())
	})

}

func TestProcessor_Context(t *testing.T) {
	p := mono.CreateProcessor()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	done := make(chan struct{})
	p.
		DoOnError(func(e error) {
			assert.True(t, reactor.IsCancelledError(e))
		}).
		DoFinally(func(signal reactor.SignalType) {
			assert.Equal(t, reactor.SignalTypeCancel, signal)
			close(done)
		}).
		Subscribe(ctx)
	<-done
}

func TestProcessor_Error(t *testing.T) {
	fakeErr := errors.New("fake error")
	p := mono.CreateProcessor()
	done := make(chan error, 1)
	p.
		DoOnError(func(e error) {
			done <- e
		}).
		Subscribe(context.Background())

	time.Sleep(100 * time.Millisecond)
	p.Error(fakeErr)
	e := <-done
	assert.Equal(t, fakeErr, e)
}

func TestOneshotProcess(t *testing.T) {
	dropped := new(int32)

	hooks.OnErrorDrop(func(e error) {
		atomic.AddInt32(dropped, 1)
	})

	const total = 10000

	var wg sync.WaitGroup
	wg.Add(total)

	c := make(chan mono.Sink, 10)

	go func() {
		var n int
		for next := range c {
			next.Success(n)
			n++
		}
	}()

	success := new(int32)

	sub := reactor.
		NewSubscriber(
			reactor.OnNext(func(v reactor.Any) error {
				atomic.AddInt32(success, 1)
				return nil
			}),
			reactor.OnComplete(func() {
				wg.Done()
			}),
			reactor.OnError(func(e error) {
				wg.Done()
			}),
		)

	doFinallyCalls := new(int32)

	for i := 0; i < total; i++ {
		m, s := mono.CreateProcessorOneshot()
		c <- s
		m.
			DoFinally(func(s reactor.SignalType) {
				atomic.AddInt32(doFinallyCalls, 1)
			}).
			SubscribeOn(scheduler.Parallel()).
			SubscribeWith(context.Background(), sub)
	}

	close(c)

	wg.Wait()

	assert.Equal(t, int32(0), atomic.LoadInt32(dropped))
	assert.Equal(t, int32(total), atomic.LoadInt32(success))
	assert.Equal(t, int32(total), atomic.LoadInt32(doFinallyCalls))
}
