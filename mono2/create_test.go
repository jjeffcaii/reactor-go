package mono

import (
	"context"
	"errors"
	"testing"

	"go.uber.org/mock/gomock"

	"github.com/jjeffcaii/reactor-go"
	"github.com/jjeffcaii/reactor-go/testdata"
)

func TestMonoCreate_SubscribeWith(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	s := testdata.NewMockSubscriber[int](ctrl)
	s.EXPECT().OnSubscribe(gomock.Any(), gomock.Any()).Do(MockRequestInfinite).Times(1)
	s.EXPECT().OnNext(gomock.Any(), gomock.Eq(1)).Times(1)
	s.EXPECT().OnComplete(gomock.Any()).Times(1)
	s.EXPECT().OnError(gomock.Any(), gomock.Any()).Times(0)

	newMonoCreate[int](func(ctx context.Context, s Sink[int]) {
		s.Success(1)
	}).SubscribeWith(context.TODO(), s)

	s = testdata.NewMockSubscriber[int](ctrl)
	s.EXPECT().OnSubscribe(gomock.Any(), gomock.Any()).Do(MockRequestInfinite).Times(1)
	s.EXPECT().OnNext(gomock.Any(), gomock.Any()).Times(0)
	s.EXPECT().OnComplete(gomock.Any()).Times(0)
	s.EXPECT().OnError(gomock.Any(), gomock.Eq(fakeErr)).Times(1)

	newMonoCreate[int](func(ctx context.Context, s Sink[int]) {
		s.Error(fakeErr)
	}).SubscribeWith(context.TODO(), s)
}

func TestMonoCreate_Panic(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	s := testdata.NewMockSubscriber[int](ctrl)
	s.EXPECT().OnSubscribe(gomock.Any(), gomock.Any()).Do(MockRequestInfinite).Times(2)
	s.EXPECT().OnNext(gomock.Any(), gomock.Any()).Times(0)
	s.EXPECT().OnError(gomock.Any(), gomock.Any()).Times(2)
	s.EXPECT().OnComplete(gomock.Any()).Times(0)

	newMonoCreate[int](func(ctx context.Context, s Sink[int]) {
		panic("fake panic")
	}).SubscribeWith(context.TODO(), s)

	newMonoCreate[int](func(ctx context.Context, s Sink[int]) {
		panic(errors.New("fake error"))
	}).SubscribeWith(context.Background(), s)
}

func TestMonoCreate_OnNextPanic(t *testing.T) {
	var cnt int
	onError := reactor.OnError[int](func(ctx context.Context, e error) {
		cnt++
	})
	m := newMonoCreate[int](func(ctx context.Context, s Sink[int]) {
		s.Success(1)
	})
	m.SubscribeWith(context.Background(),
		reactor.NewSubscriber[int](
			reactor.OnNext(func(ctx context.Context, v int) error {
				panic("fake panic")
			}),
			onError,
		))
	m.SubscribeWith(context.Background(),
		reactor.NewSubscriber[int](
			reactor.OnNext(func(ctx context.Context, v int) error {
				panic(errors.New("fake error"))
			}),
			onError,
		))
}

func TestMonoCreate_Cancel(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	s := testdata.NewMockSubscriber[int](ctrl)
	s.EXPECT().OnSubscribe(gomock.Any(), gomock.Any()).
		Do(func(ctx context.Context, su reactor.Subscription) {
			su.Cancel(ctx)
		}).
		Times(1)
	s.EXPECT().OnNext(gomock.Any(), gomock.Any()).Times(0)
	s.EXPECT().OnComplete(gomock.Any()).Times(0)
	s.EXPECT().OnError(gomock.Any(), gomock.Eq(reactor.ErrSubscribeCancelled)).Times(1)

	newMonoCreate[int](func(ctx context.Context, s Sink[int]) {
		s.Success(1)
	}).SubscribeWith(context.TODO(), s)
}
