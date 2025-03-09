package subscribers

import (
	"context"
	"errors"
	"testing"

	"go.uber.org/mock/gomock"

	"github.com/jjeffcaii/reactor-go"
	"github.com/jjeffcaii/reactor-go/testdata"
)

func TestSwitchIfEmptySubscriber(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	su := testdata.NewMockSubscription(ctrl)
	su.EXPECT().Request(gomock.Any(), gomock.Any()).AnyTimes()
	su.EXPECT().Cancel(gomock.Any()).AnyTimes()

	const alternativeValue = 42

	pub := testdata.NewMockRawPublisher[int](ctrl)
	subscribeWith := func(ctx context.Context, sub reactor.Subscriber[int]) {
		sub.OnSubscribe(ctx, su)
		sub.OnNext(ctx, alternativeValue)
		sub.OnComplete(ctx)
	}
	pub.EXPECT().SubscribeWith(gomock.Any(), gomock.Any()).Do(subscribeWith).Times(1)

	actual := testdata.NewMockSubscriber[int](ctrl)
	actual.EXPECT().OnSubscribe(gomock.Any(), gomock.Any()).Times(0)
	actual.EXPECT().OnNext(gomock.Any(), gomock.Eq(alternativeValue)).Times(1)
	actual.EXPECT().OnComplete(gomock.Any()).Times(1)
	actual.EXPECT().OnError(gomock.Any(), gomock.Any()).Times(0)

	s := NewSwitchIfEmptySubscriber[int](pub, actual)
	s.OnSubscribe(context.TODO(), su)
	s.OnComplete(context.TODO())

	const value = 1024

	actual = testdata.NewMockSubscriber[int](ctrl)
	actual.EXPECT().OnSubscribe(gomock.Any(), gomock.Any()).Times(0)
	actual.EXPECT().OnNext(gomock.Any(), gomock.Eq(value)).Times(1)
	actual.EXPECT().OnComplete(gomock.Any()).Times(1)
	actual.EXPECT().OnError(gomock.Any(), gomock.Any()).Times(0)

	s = NewSwitchIfEmptySubscriber[int](pub, actual)
	s.OnSubscribe(context.TODO(), su)
	s.OnNext(context.TODO(), value)
	s.OnComplete(context.TODO())

	fakeErr := errors.New("fake error")
	actual = testdata.NewMockSubscriber[int](ctrl)
	actual.EXPECT().OnSubscribe(gomock.Any(), gomock.Any()).Times(0)
	actual.EXPECT().OnNext(gomock.Any(), gomock.Any()).Times(0)
	actual.EXPECT().OnComplete(gomock.Any()).Times(0)
	actual.EXPECT().OnError(gomock.Any(), gomock.Eq(fakeErr)).Times(1)

	s = NewSwitchIfEmptySubscriber[int](pub, actual)
	s.OnSubscribe(context.TODO(), su)
	s.OnError(context.TODO(), fakeErr)
}
