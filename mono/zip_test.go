package mono_test

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/jjeffcaii/reactor-go"
	"github.com/jjeffcaii/reactor-go/mono"
	"github.com/jjeffcaii/reactor-go/tuple"
	"github.com/stretchr/testify/assert"
)

func TestZip(t *testing.T) {
	begin := time.Now()
	v, err := mono.Zip(mono.Delay(100*time.Millisecond), mono.Delay(200*time.Millisecond)).Block(context.Background())
	assert.NoError(t, err)
	t.Log("cost:", time.Since(begin))
	tu, ok := v.(tuple.Tuple)
	assert.True(t, ok)
	assert.Equal(t, int64(0), tu.GetValue(0))
	assert.Equal(t, int64(0), tu.GetValue(1))
}

func TestZip_Bad(t *testing.T) {
	assert.Panics(t, func() {
		mono.Zip(nil, nil)
	})
	assert.Panics(t, func() {
		mono.Zip(mono.Just(1), nil)
	})
	assert.Panics(t, func() {
		mono.Zip(mono.Just(1), mono.Just(2), nil)
	})
	assert.Panics(t, func() {
		mono.Zip(mono.Just(1), mono.Just(2), nil, mono.Just(3))
	})
	assert.Panics(t, func() {
		mono.Zip(mono.Just(1), mono.Just(2), mono.Just(3), nil)
	})
}

func TestZipCombine_Bad(t *testing.T) {
	assert.Panics(t, func() {
		mono.ZipCombine(nil, nil)
	})
	assert.Panics(t, func() {
		mono.ZipCombine(nil, nil, mono.Just(1))
	})
	assert.Panics(t, func() {
		mono.ZipCombine(nil, nil, mono.Just(1), mono.Just(2), nil)
	})
	assert.Panics(t, func() {
		mono.ZipCombine(nil, nil, nil)
	})
	assert.Panics(t, func() {
		mono.ZipCombine(nil, nil, mono.Just(1), nil)
	})
}

func TestZipWithError(t *testing.T) {
	var (
		fakeErr1 = errors.New("fake error 1")
		fakeErr2 = errors.New("fake error 2")
	)
	v, err := mono.Zip(mono.Error(fakeErr1), mono.Error(fakeErr2)).Block(context.Background())
	assert.NoError(t, err, "should not return error")
	tu := v.(tuple.Tuple)
	_, e1 := tu.First()
	_, e2 := tu.Second()
	assert.Equal(t, fakeErr1, e1)
	assert.Equal(t, fakeErr2, e2)
}

func TestZipWith_Mixin(t *testing.T) {
	val, err := mono.Zip(mono.Just(1), mono.Just(2), mono.Error(fakeErr)).Block(context.Background())
	assert.NoError(t, err)
	tup := val.(tuple.Tuple)
	assert.Equal(t, 3, tup.Len())
}

func TestZip_Map(t *testing.T) {
	v, err := mono.Zip(mono.Just(1), mono.Just(2), mono.Just(3)).
		Map(func(any reactor.Any) (reactor.Any, error) {
			tu := any.(tuple.Tuple)
			var sum int
			for i := 0; i < tu.Len(); i++ {
				sum += tu.GetValue(i).(int)
			}
			return sum, nil
		}).
		Block(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, 6, v)
}

func TestZip_Cancel(t *testing.T) {
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

func TestZipCombine(t *testing.T) {
	cmb := func(items ...*reactor.Item) (reactor.Any, error) {
		values := make([]reactor.Any, len(items))
		for i := 0; i < len(items); i++ {
			values[i] = items[i].V
		}
		return values, nil
	}
	v, err := mono.ZipCombine(cmb, nil, mono.Just(1), mono.Just(2)).Block(context.Background())
	assert.NoError(t, err, "should not return error")
	values := v.([]reactor.Any)
	assert.Len(t, values, 2)
	assert.Equal(t, 1, values[0], "incorrect first value")
	assert.Equal(t, 2, values[1], "incorrect second value")
}

func TestZip_context(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := mono.Zip(mono.Delay(200*time.Millisecond), mono.Delay(100*time.Millisecond)).Block(ctx)
	assert.Error(t, err)
	assert.True(t, reactor.IsCancelledError(err))
}

func TestZip_EdgeCase(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	sub := mono.NewMockSubscriber(ctrl)

	sub.EXPECT().OnSubscribe(gomock.Any(), gomock.Any()).
		Do(func(ctx context.Context, su reactor.Subscription) {
			su.Request(reactor.RequestInfinite)
		}).
		Times(1)
	sub.EXPECT().OnNext(gomock.Any()).Times(0)
	sub.EXPECT().OnError(gomock.Any()).Times(1)
	sub.EXPECT().OnComplete().Times(0)

	mono.Zip(mono.JustOneshot("1"), mono.JustOneshot("2")).
		FlatMap(func(any reactor.Any) mono.Mono {
			if any != nil {
				return mono.Zip(mono.JustOneshot("333"), mono.JustOneshot("44444444")).
					Map(func(any reactor.Any) (reactor.Any, error) {
						panic("fake panic")
					})
			}
			return mono.JustOneshot("dddd")
		}).
		SubscribeWith(context.Background(), sub)
}

func TestZipWith(t *testing.T) {
	cmb := func(values ...*reactor.Item) (reactor.Any, error) {
		var sum int
		for i := 0; i < len(values); i++ {
			sum += values[i].V.(int)
		}
		return sum, nil
	}
	v, err := mono.Just(1).
		ZipCombineWith(mono.Just(2), cmb).
		ZipCombineWith(mono.Just(3), cmb).
		Map(func(any reactor.Any) (reactor.Any, error) {
			return any.(int) * 2, nil
		}).
		Block(context.Background())
	assert.NoError(t, err, "should not returns with error")
	assert.Equal(t, (1+2+3)*2, v, "should be (1+2+3)*2")
}

func TestZip_Empty(t *testing.T) {
	res, err := mono.Zip(mono.Just(1), mono.Empty()).Block(context.Background())
	assert.NoError(t, err)
	tu := res.(tuple.Tuple)
	assert.Equal(t, 1, tu.GetValue(0))
	assert.Nil(t, tu.GetValue(1))

	res, err = mono.
		Zip(
			mono.Just(1).
				Filter(func(any reactor.Any) bool {
					return any.(int) > 1
				}),
			mono.Just(2),
		).
		Block(context.Background())
	assert.NoError(t, err)
	tu = res.(tuple.Tuple)
	assert.Nil(t, tu.GetValue(0))
	assert.Equal(t, 2, tu.GetValue(1))
}

func TestZipCombineWithItemHandler(t *testing.T) {
	var cnt int
	handler := func(item *reactor.Item) {
		cnt++
	}
	mono.ZipCombine(nil, handler, mono.Just(1), mono.Error(fakeErr)).Subscribe(context.Background())
	assert.Equal(t, 2, cnt)
}

func TestZipCombine_Panic(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	s := mono.NewMockSubscriber(ctrl)
	onSub := func(ctx context.Context, su reactor.Subscription) {
		su.Request(-1)
		su.Request(0)
		su.Request(1)
		su.Request(reactor.RequestInfinite)
	}
	s.EXPECT().OnSubscribe(gomock.Any(), gomock.Any()).Do(onSub).Times(1)
	s.EXPECT().OnError(gomock.Any()).Times(1)
	s.EXPECT().OnNext(gomock.Any()).Times(0)
	s.EXPECT().OnComplete().Times(0)

	cmb := func(values ...*reactor.Item) (reactor.Any, error) {
		panic("fake panic")
	}
	mono.ZipCombine(cmb, nil, mono.Just(1), mono.Just(2)).SubscribeWith(context.Background(), s)
}

func TestZipOneshot(t *testing.T) {
	res, err := mono.ZipOneshot(mono.JustOneshot(1), mono.JustOneshot(2).Timeout(10*time.Millisecond)).Block(context.Background())
	assert.NoError(t, err)
	tu := res.(tuple.Tuple)
	assert.Equal(t, 1, tu.GetValue(0))
	assert.Equal(t, 2, tu.GetValue(1))
}
