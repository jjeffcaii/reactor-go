package mono

import (
	"context"
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"

	"github.com/jjeffcaii/reactor-go"
	"github.com/jjeffcaii/reactor-go/testdata"
)

var fakeErr = errors.New("fake error")

func TestMonoMap_SubscribeWith(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	const value = 42
	tf := func(ctx context.Context, next int) (int, error) {
		return next * 2, nil
	}

	// happy path with value
	s := testdata.NewMockSubscriber[int](ctrl)
	s.EXPECT().OnSubscribe(gomock.Any(), gomock.Any()).Do(MockRequestInfinite).Times(1)
	s.EXPECT().OnError(gomock.Any(), gomock.Any()).Times(0)
	s.EXPECT().OnNext(gomock.Any(), gomock.Eq(value*2)).Times(1)
	s.EXPECT().OnComplete(gomock.Any()).Times(1)

	newMonoMap(newMonoJust(value), tf).SubscribeWith(context.TODO(), s)

	// happy path with error
	fakeErr := errors.New("fake error")
	s = testdata.NewMockSubscriber[int](ctrl)
	s.EXPECT().OnSubscribe(gomock.Any(), gomock.Any()).Do(MockRequestInfinite).Times(1)
	s.EXPECT().OnError(gomock.Any(), gomock.Eq(fakeErr)).Times(1)
	s.EXPECT().OnNext(gomock.Any(), gomock.Any()).Times(0)
	s.EXPECT().OnComplete(gomock.Any()).Times(0)

	newMonoMap[int, int](newMonoError[int](fakeErr), tf).SubscribeWith(context.Background(), s)
}

func TestMonoMap_Panic(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	s := testdata.NewMockSubscriber[int](ctrl)
	s.EXPECT().OnSubscribe(gomock.Any(), gomock.Any()).Do(MockRequestInfinite).Times(1)
	s.EXPECT().OnError(gomock.Any(), gomock.Any()).Times(1)
	s.EXPECT().OnNext(gomock.Any(), gomock.Any()).Times(0)
	s.EXPECT().OnComplete(gomock.Any()).Times(0)

	origin := newMonoJust(1)
	tf := func(ctx context.Context, next int) (int, error) {
		panic("fake panic")
	}
	newMonoMap[int, int](origin, tf).SubscribeWith(context.TODO(), s)

	s = testdata.NewMockSubscriber[int](ctrl)
	s.EXPECT().
		OnSubscribe(gomock.Any(), gomock.Any()).
		Do(MockRequestInfinite).
		Times(1)
	s.EXPECT().
		OnError(gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, err error) error {
			assert.Equal(t, fakeErr, errors.Cause(err))
			return nil
		}).
		Times(1)
	s.EXPECT().OnNext(gomock.Any(), gomock.Any()).Times(0)
	s.EXPECT().OnComplete(gomock.Any()).Times(0)

	newMonoMap[int, int](newMonoJust(1), func(ctx context.Context, next int) (int, error) {
		panic(fakeErr)
	}).SubscribeWith(context.TODO(), s)
}

func TestMonoMap_Cancel(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	onSubscribe := func(ctx context.Context, su reactor.Subscription) {
		su.Cancel(ctx)
	}

	s := testdata.NewMockSubscriber[int](ctrl)
	s.EXPECT().OnSubscribe(gomock.Any(), gomock.Any()).Do(onSubscribe).Times(1)
	s.EXPECT().OnError(gomock.Any(), gomock.Eq(reactor.ErrSubscribeCancelled)).Times(1)
	s.EXPECT().OnNext(gomock.Any(), gomock.Any()).Times(0)
	s.EXPECT().OnComplete(gomock.Any()).Times(0)

	origin := newMonoJust(1)
	tf := func(ctx context.Context, next int) (int, error) {
		return next * 2, nil
	}
	newMonoMap[int, int](origin, tf).SubscribeWith(context.TODO(), s)
}
