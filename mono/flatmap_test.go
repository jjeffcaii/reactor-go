package mono_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/jjeffcaii/reactor-go"
	"github.com/jjeffcaii/reactor-go/mono"
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
