package mono_test

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"github.com/jjeffcaii/reactor-go"
	"github.com/jjeffcaii/reactor-go/mono"
	"github.com/jjeffcaii/reactor-go/tuple"
	"github.com/stretchr/testify/assert"
)

func TestZip(t *testing.T) {
	begin := time.Now()
	val, err := mono.Zip(mono.Delay(100*time.Millisecond), mono.Delay(200*time.Millisecond)).Block(context.Background())
	assert.NoError(t, err)
	t.Log("cost:", time.Since(begin))
	tup := val.(tuple.Tuple)
	for i := 0; i < tup.Len(); i++ {
		v, e := tup.Get(i)
		assert.NoError(t, e)
		assert.Equal(t, int64(0), v)
	}

	_, err = tup.Get(-1)
	assert.True(t, tuple.IsIndexOutOfBoundsError(err))

	_, err = tup.Get(tup.Len())
	assert.True(t, tuple.IsIndexOutOfBoundsError(err))
}

func TestZipWithError(t *testing.T) {
	var (
		fakeErr1 = errors.New("fake error 1")
		fakeErr2 = errors.New("fake error 2")
	)
	val, err := mono.Zip(mono.Error(fakeErr1), mono.Error(fakeErr2)).Block(context.Background())
	assert.NoError(t, err)
	tup := val.(tuple.Tuple)
	_, err = tup.First()
	assert.Equal(t, fakeErr1, err)
	_, err = tup.Second()
	assert.Equal(t, fakeErr2, err)
}

func TestZipWithValueAndError(t *testing.T) {
	val, err := mono.Zip(mono.Just(1), mono.Just(2), mono.Error(fakeErr)).Block(context.Background())
	assert.NoError(t, err)
	tup := val.(tuple.Tuple)
	assert.Equal(t, 3, tup.Len())
}

func TestZipMap(t *testing.T) {
	v, err := mono.Zip(mono.Just(1), mono.Just(2), mono.Just(3)).
		Map(func(any reactor.Any) (reactor.Any, error) {
			tup := any.(tuple.Tuple)
			var sum int
			tup.ForEach(func(v reactor.Any, e error) (ok bool) {
				sum += v.(int)
				ok = true
				return
			})
			return sum, nil
		}).
		Block(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, 6, v)
}

func TestZipCancel(t *testing.T) {
	callDoOnCancel := new(int32)
	done := make(chan struct{})
	mono.Zip(
		mono.Delay(1*time.Second).DoOnCancel(func() {
			t.Log("cancelled1")
			atomic.AddInt32(callDoOnCancel, 1)
		}),
		mono.Delay(2*time.Second).DoOnCancel(func() {
			t.Log("cancelled2")
			atomic.AddInt32(callDoOnCancel, 1)
		})).
		DoFinally(func(s reactor.SignalType) {
			close(done)
		}).
		DoOnCancel(func() {
			t.Log("cancelled")
			atomic.AddInt32(callDoOnCancel, 1)
		}).
		Subscribe(context.Background(),
			reactor.OnSubscribe(func(ctx context.Context, su reactor.Subscription) {
				time.Sleep(100 * time.Millisecond)
				su.Cancel()
			}),
			reactor.OnNext(func(v reactor.Any) error {
				t.Log("next:", v)
				return nil
			}),
		)
	<-done
	assert.Equal(t, int32(3), atomic.LoadInt32(callDoOnCancel))
}

func TestZipAll(t *testing.T) {
	assert.Panics(t, func() {
		mono.ZipAll()
	}, "should panic without sources")
	v, err := mono.ZipAll(mono.Just(1)).Block(context.Background())
	assert.NoError(t, err, "should not return error")
	v, err = v.(tuple.Tuple).First()
	assert.NoError(t, err, "should not return error")
	assert.Equal(t, 1, v, "should be same value")
}

func TestZip_context(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := mono.Zip(mono.Delay(200*time.Millisecond), mono.Delay(100*time.Millisecond)).Block(ctx)
	assert.Error(t, err)
	assert.True(t, reactor.IsCancelledError(err))
}

func TestZip_EdgeCase(t *testing.T) {

	var (
		nextCnt     = new(int32)
		completeCnt = new(int32)
		errorCnt    = new(int32)
	)

	mono.Zip(mono.JustOneshot("1"), mono.JustOneshot("2")).
		FlatMap(func(any reactor.Any) mono.Mono {
			if any != nil {
				return mono.Zip(mono.JustOneshot("333"), mono.JustOneshot("44444444")).
					Filter(func(any reactor.Any) bool {
						panic("fake panic")
					}).
					Map(func(any reactor.Any) (reactor.Any, error) {
						//此处会造成无任何返回值，直接抛出异常
						panic("ddddddd")
						return any, nil
					})
			}
			return mono.JustOneshot("dddd")
		}).
		Subscribe(context.Background(),
			reactor.OnNext(func(v reactor.Any) error {
				atomic.AddInt32(nextCnt, 1)
				return nil
			}),
			reactor.OnError(func(e error) {
				atomic.AddInt32(errorCnt, 1)
				t.Logf("%v", e)
			}),
			reactor.OnComplete(func() {
				atomic.AddInt32(completeCnt, 1)
			}),
		)

	assert.Equal(t, int32(0), atomic.LoadInt32(nextCnt), "next count should be zero")
	assert.Equal(t, int32(1), atomic.LoadInt32(errorCnt), "error count should be 1")
	assert.Equal(t, int32(0), atomic.LoadInt32(completeCnt), "complete count should be zero")
}
