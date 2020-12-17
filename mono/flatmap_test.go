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
	"github.com/jjeffcaii/reactor-go/scheduler"
	"github.com/stretchr/testify/assert"
)

func TestFlatMap(t *testing.T) {
	const fakeItem = "fake item"
	actual, err := mono.Just(123).
		FlatMap(func(_ reactor.Any) mono.Mono {
			return mono.Delay(300 * time.Millisecond).
				Map(func(_ reactor.Any) (reactor.Any, error) {
					return fakeItem, nil
				})
		}).
		Block(context.Background())
	assert.NoError(t, err, "should not return error")
	assert.Equal(t, fakeItem, actual)
}

func TestFlatMap_Panic(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	s := mono.NewMockSubscriber(ctrl)
	s.EXPECT().OnSubscribe(gomock.Any(), gomock.Any()).Do(mono.MockRequestInfinite).Times(1)
	s.EXPECT().OnNext(gomock.Any()).Times(0)
	s.EXPECT().OnError(gomock.Any()).Times(1)
	s.EXPECT().OnComplete().Times(0)

	transform := func(any reactor.Any) mono.Mono {
		panic("fake panic")
	}
	mono.Just(1).FlatMap(transform).SubscribeWith(context.Background(), s)
}

func TestFlatMap_Error(t *testing.T) {
	mono.ErrorOneshot(fakeErr).
		FlatMap(func(any reactor.Any) mono.Mono {
			assert.Fail(t, "unreachable")
			return mono.Error(errors.New("unreachable"))
		}).
		Subscribe(context.Background(), reactor.OnNext(func(v reactor.Any) error {
			assert.Fail(t, "unreachable")
			return nil
		}), reactor.OnError(func(e error) {
			assert.Equal(t, fakeErr, e)
		}))

}

func TestFlatMap_MultipleEmit(t *testing.T) {
	res, err := mono.
		CreateOneshot(func(ctx context.Context, s mono.Sink) {
			s.Success(1)
			s.Success("duplicated")
			s.Error(fakeErr)
		}).
		FlatMap(func(any reactor.Any) mono.Mono {
			return mono.Just(any)
		}).
		Block(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, 1, res)
}

func TestFlatMapSubscriber_OnComplete(t *testing.T) {
	completes := new(int32)
	res, err := mono.Just(1).
		FlatMap(func(any reactor.Any) mono.Mono {
			return mono.Create(func(ctx context.Context, s mono.Sink) {
				s.Success(2)
			}).SubscribeOn(scheduler.Parallel())
		}).
		DoOnComplete(func() {
			atomic.AddInt32(completes, 1)
		}).
		Block(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, 2, res, "bad result")
	assert.Equal(t, int32(1), atomic.LoadInt32(completes), "completes should be 1")

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	s := mono.NewMockSubscriber(ctrl)
	s.EXPECT().OnComplete().Times(1)
	s.EXPECT().OnNext(gomock.Eq(2)).Times(1)
	s.EXPECT().OnError(gomock.Any()).Times(0)
	s.EXPECT().OnSubscribe(gomock.Any(), gomock.Any()).Do(mono.MockRequestInfinite).Times(1)

	mono.Just(1).
		FlatMap(func(any reactor.Any) mono.Mono {
			return mono.Just(2)
		}).
		SubscribeWith(context.Background(), s)
}

func TestFlatMap_EmptySource(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	s := mono.NewMockSubscriber(ctrl)
	s.EXPECT().OnComplete().Times(1)
	s.EXPECT().OnNext(gomock.Any()).Times(0)
	s.EXPECT().OnError(gomock.Any()).Times(0)
	s.EXPECT().OnSubscribe(gomock.Any(), gomock.Any()).Do(mono.MockRequestInfinite).Times(1)

	mono.Empty().
		FlatMap(func(any reactor.Any) mono.Mono {
			return mono.Just(1)
		}).
		SubscribeWith(context.Background(), s)

	s = mono.NewMockSubscriber(ctrl)
	s.EXPECT().OnComplete().Times(1)
	s.EXPECT().OnNext(gomock.Any()).Times(0)
	s.EXPECT().OnError(gomock.Any()).Times(0)
	s.EXPECT().OnSubscribe(gomock.Any(), gomock.Any()).Do(mono.MockRequestInfinite).Times(1)

	mono.Just(1).
		FlatMap(func(value reactor.Any) mono.Mono {
			assert.Equal(t, 1, value)
			return mono.Empty()
		}).
		SubscribeWith(context.Background(), s)
}
