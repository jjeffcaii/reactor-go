package mono

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/jjeffcaii/reactor-go"
)

func MockRequestInfinite(_ context.Context, su reactor.Subscription) {
	su.Request(reactor.RequestInfinite)
}

func TestMonoJust_SubscribeWith(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	sub := NewMockSubscriber(ctrl)
	sub.EXPECT().OnSubscribe(gomock.Any(), gomock.Any()).Do(MockRequestInfinite).Times(1)
	sub.EXPECT().OnNext(gomock.Eq(1)).Times(1)
	sub.EXPECT().OnComplete().Times(1)
	sub.EXPECT().OnError(gomock.Any()).Times(0)
	m := newMonoJust(1)
	m.SubscribeWith(context.Background(), sub)
	//ctx, cancel := context.WithCancel(context.Background())
	//cancel()
	//m.SubscribeWith(ctx, nil)
	//sub.(reactor.Subscriber).OnNext(1)
}

func TestMonoJust_SubscribeWith_Context(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	sub := NewMockSubscriber(ctrl)
	sub.EXPECT().OnSubscribe(gomock.Any(), gomock.Any()).Times(0)
	sub.EXPECT().OnNext(gomock.Any()).Times(0)
	sub.EXPECT().OnComplete().Times(0)
	sub.EXPECT().OnError(gomock.Any()).Times(1)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	m := newMonoJust(1)
	m.SubscribeWith(ctx, sub)
}

func TestMonoJust_Cancel(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	sub := NewMockSubscriber(ctrl)

	onSubscribe := func(ctx context.Context, su reactor.Subscription) {
		su.Cancel()
	}
	sub.EXPECT().OnSubscribe(gomock.Any(), gomock.Any()).Do(onSubscribe).Times(1)
	sub.EXPECT().OnNext(gomock.Any()).Times(0)
	sub.EXPECT().OnComplete().Times(0)
	sub.EXPECT().OnError(gomock.Eq(reactor.ErrSubscribeCancelled)).Times(1)

	m := newMonoJust(1)
	m.SubscribeWith(context.Background(), sub)
}

func TestMonoJust_Request(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	sub := NewMockSubscriber(ctrl)

	onSubscribe := func(ctx context.Context, su reactor.Subscription) {
		su.Request(1)
		su.Request(1)
	}
	sub.EXPECT().OnSubscribe(gomock.Any(), gomock.Any()).Do(onSubscribe).Times(1)
	sub.EXPECT().OnNext(gomock.Any()).Times(1)
	sub.EXPECT().OnComplete().Times(1)
	sub.EXPECT().OnError(gomock.Any()).Times(0)

	m := newMonoJust(1)
	m.SubscribeWith(context.Background(), sub)
}
