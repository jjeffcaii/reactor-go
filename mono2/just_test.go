package mono

import (
	"context"
	"testing"

	"go.uber.org/mock/gomock"

	"github.com/jjeffcaii/reactor-go"
	"github.com/jjeffcaii/reactor-go/testdata"
)

func MockRequestInfinite(ctx context.Context, su reactor.Subscription) {
	su.Request(ctx, reactor.RequestInfinite)
}

func TestMonoJust_SubscribeWith(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	sub := testdata.NewMockSubscriber[int](ctrl)
	sub.EXPECT().OnSubscribe(gomock.Any(), gomock.Any()).Do(MockRequestInfinite).Times(1)
	sub.EXPECT().OnNext(gomock.Any(), gomock.Eq(1)).Times(1)
	sub.EXPECT().OnComplete(gomock.Any()).Times(1)
	sub.EXPECT().OnError(gomock.Any(), gomock.Any()).Times(0)

	m := newMonoJust[int](1)
	m.SubscribeWith(context.Background(), sub)
}

func TestMonoJust_SubscribeWith_Context(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	sub := testdata.NewMockSubscriber[int](ctrl)
	sub.EXPECT().OnSubscribe(gomock.Any(), gomock.Any()).Times(0)
	sub.EXPECT().OnNext(gomock.Any(), gomock.Any()).Times(0)
	sub.EXPECT().OnComplete(gomock.Any()).Times(0)
	sub.EXPECT().OnError(gomock.Any(), gomock.Any()).Times(1)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	m := newMonoJust[int](1)
	m.SubscribeWith(ctx, sub)
}

func TestMonoJust_Cancel(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	sub := testdata.NewMockSubscriber[int](ctrl)
	sub.EXPECT().
		OnSubscribe(gomock.Any(), gomock.Any()).
		Do(func(ctx context.Context, su reactor.Subscription) {
			su.Cancel(ctx)
		}).
		Times(1)
	sub.EXPECT().OnNext(gomock.Any(), gomock.Any()).Times(0)
	sub.EXPECT().OnComplete(gomock.Any()).Times(0)
	sub.EXPECT().OnError(gomock.Any(), gomock.Eq(reactor.ErrSubscribeCancelled)).Times(1)

	m := newMonoJust(1)
	m.SubscribeWith(context.Background(), sub)
}

func TestMonoJust_Request(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	sub := testdata.NewMockSubscriber[int](ctrl)
	onSubscribe := func(ctx context.Context, su reactor.Subscription) {
		su.Request(ctx, 1)
		su.Request(ctx, 1)
	}
	sub.EXPECT().OnSubscribe(gomock.Any(), gomock.Any()).Do(onSubscribe).Times(1)
	sub.EXPECT().OnNext(gomock.Any(), gomock.Any()).Times(1)
	sub.EXPECT().OnComplete(gomock.Any()).Times(1)
	sub.EXPECT().OnError(gomock.Any(), gomock.Any()).Times(0)

	m := newMonoJust[int](1)
	m.SubscribeWith(context.Background(), sub)
}
