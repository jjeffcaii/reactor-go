package subscribers

import (
	"context"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/jjeffcaii/reactor-go"
)

func TestSwitchIfEmptySubscriber(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	su := NewMockSubscription(ctrl)
	su.EXPECT().Request(gomock.Any()).AnyTimes()
	su.EXPECT().Cancel().AnyTimes()

	const alternativeValue = 42

	pub := NewMockRawPublisher(ctrl)
	subscribeWith := func(ctx context.Context, sub reactor.Subscriber) {
		sub.OnSubscribe(ctx, su)
		sub.OnNext(alternativeValue)
		sub.OnComplete()
	}
	pub.EXPECT().SubscribeWith(gomock.Any(), gomock.Any()).Do(subscribeWith).Times(1)

	actual := NewMockSubscriber(ctrl)
	actual.EXPECT().OnSubscribe(gomock.Any(), gomock.Any()).Times(0)
	actual.EXPECT().OnNext(gomock.Eq(alternativeValue)).Times(1)
	actual.EXPECT().OnComplete().Times(1)
	actual.EXPECT().OnError(gomock.Any()).Times(0)

	s := NewSwitchIfEmptySubscriber(pub, actual)
	s.OnSubscribe(context.Background(), su)
	s.OnComplete()

	const value = 1024

	actual = NewMockSubscriber(ctrl)
	actual.EXPECT().OnSubscribe(gomock.Any(), gomock.Any()).Times(0)
	actual.EXPECT().OnNext(gomock.Eq(value)).Times(1)
	actual.EXPECT().OnComplete().Times(1)
	actual.EXPECT().OnError(gomock.Any()).Times(0)

	s = NewSwitchIfEmptySubscriber(pub, actual)
	s.OnSubscribe(context.Background(), su)
	s.OnNext(value)
	s.OnComplete()

	fakeErr := errors.New("fake error")
	actual = NewMockSubscriber(ctrl)
	actual.EXPECT().OnSubscribe(gomock.Any(), gomock.Any()).Times(0)
	actual.EXPECT().OnNext(gomock.Any()).Times(0)
	actual.EXPECT().OnComplete().Times(0)
	actual.EXPECT().OnError(gomock.Eq(fakeErr)).Times(1)

	s = NewSwitchIfEmptySubscriber(pub, actual)
	s.OnSubscribe(context.Background(), su)
	s.OnError(fakeErr)
}
