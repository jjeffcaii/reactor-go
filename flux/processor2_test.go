package flux_test

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/jjeffcaii/reactor-go"
	"github.com/jjeffcaii/reactor-go/flux"
	"github.com/jjeffcaii/reactor-go/hooks"
	"github.com/jjeffcaii/reactor-go/scheduler"
	"github.com/stretchr/testify/assert"
)

var _fakeProcessorItems = []string{"foo", "bar", "qux"}

func TestProcessor(t *testing.T) {
	f, s, d := flux.NewProcessor(scheduler.Immediate())
	defer d.Dispose()

	go func(sink flux.Sink) {
		for i := 0; i < len(_fakeProcessorItems); i++ {
			time.Sleep(10 * time.Millisecond)
			sink.Next(_fakeProcessorItems[i])
		}
		sink.Complete()
	}(s)

	onCompleteCalls := new(int32)

	var results []string
	onNext := func(v reactor.Any) error {
		results = append(results, v.(string))
		return nil
	}
	onComplete := func() {
		atomic.AddInt32(onCompleteCalls, 1)
	}
	f.DoOnNext(onNext).DoOnComplete(onComplete).Subscribe(context.Background())

	assert.Equal(t, results, _fakeProcessorItems)
	assert.Equal(t, int32(1), atomic.LoadInt32(onCompleteCalls))
}

func TestProcessor_RequestN(t *testing.T) {
	p, s, d := flux.NewProcessor(scheduler.Parallel())
	defer d.Dispose()

	go func() {
		for _, v := range _fakeProcessorItems {
			s.Next(v)
		}
		s.Complete()
	}()

	done := make(chan struct{})
	requested := new(int32)
	var results []string
	var su reactor.Subscription
	onNext := func(v reactor.Any) error {
		t.Log("next:", v)
		results = append(results, v.(string))
		su.Request(1)
		return nil
	}
	onRequest := func(n int) {
		atomic.AddInt32(requested, int32(n))
	}
	onSubscribe := reactor.OnSubscribe(func(ctx context.Context, su2 reactor.Subscription) {
		su = su2
		su.Request(1)
	})
	onComplete := reactor.OnComplete(func() {
		close(done)
	})
	p.DoOnNext(onNext).DoOnRequest(onRequest).Subscribe(context.Background(), onSubscribe, onComplete)

	<-done

	assert.Equal(t, int32(len(_fakeProcessorItems)+1), atomic.LoadInt32(requested))
	assert.Equal(t, _fakeProcessorItems, results)
}

func TestProcessor_Cancel(t *testing.T) {
	p, s, d := flux.NewProcessor(scheduler.Parallel())
	defer d.Dispose()

	var dropped []string

	hooks.OnNextDrop(func(v reactor.Any) {
		dropped = append(dropped, v.(string))
	})

	go func() {
		for _, v := range _fakeProcessorItems {
			s.Next(v)
		}
		s.Complete()
	}()
	var su reactor.Subscription

	done := make(chan struct{})

	var results []string
	p.DoOnNext(func(v reactor.Any) error {
		results = append(results, v.(string))
		t.Log("next:", v)
		su.Cancel()
		return nil
	}).Subscribe(context.Background(),
		reactor.OnSubscribe(func(ctx context.Context, suu reactor.Subscription) {
			su = suu
			su.Request(1)
		}),
		reactor.OnComplete(func() {
			close(done)
		}),
		reactor.OnError(func(e error) {
			close(done)
		}),
	)

	<-done
	assert.Equal(t, _fakeProcessorItems[:1], results, "result doesn't match")
	assert.Equal(t, _fakeProcessorItems[1:], dropped, "dropped doesn't match")
}
