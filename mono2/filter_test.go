package mono

import (
	"context"
	"testing"

	"go.uber.org/mock/gomock"

	"github.com/jjeffcaii/reactor-go/testdata"
)

func TestMonoFilter(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	sub := testdata.NewMockSubscriber[int](ctrl)

	sub.EXPECT().OnNext(gomock.Any(), gomock.Eq(42)).Times(1)
	sub.EXPECT().OnComplete(gomock.Any()).Times(1)
	sub.EXPECT().OnError(gomock.Any(), gomock.Any()).Times(0)
	sub.EXPECT().OnSubscribe(gomock.Any(), gomock.Any()).Do(MockRequestInfinite).Times(1)

	predicate := func(ctx context.Context, next int) bool {
		return next > 0
	}

	newMonoFilter(newMonoJust(42), predicate).SubscribeWith(context.TODO(), sub)

	sub = testdata.NewMockSubscriber[int](ctrl)
	sub.EXPECT().OnNext(gomock.Any(), gomock.Any()).Times(0)
	sub.EXPECT().OnComplete(gomock.Any()).Times(1)
	sub.EXPECT().OnError(gomock.Any(), gomock.Any()).Times(0)
	sub.EXPECT().OnSubscribe(gomock.Any(), gomock.Any()).Do(MockRequestInfinite).Times(1)

	newMonoFilter(newMonoJust(-42), predicate).SubscribeWith(context.TODO(), sub)

	sub = testdata.NewMockSubscriber[int](ctrl)
	sub.EXPECT().OnNext(gomock.Any(), gomock.Any()).Times(0)
	sub.EXPECT().OnComplete(gomock.Any()).Times(1)
	sub.EXPECT().OnError(gomock.Any(), gomock.Any()).Times(0)
	sub.EXPECT().OnSubscribe(gomock.Any(), gomock.Any()).Do(MockRequestInfinite).Times(1)

	empty := newMonoEmpty[int]()
	predicate = func(ctx context.Context, next int) bool {
		return true
	}
	newMonoFilter[int](empty, predicate).SubscribeWith(context.TODO(), sub)
}

func TestMonoFilter_Panic(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	sub := testdata.NewMockSubscriber[int](ctrl)

	sub.EXPECT().OnNext(gomock.Any(), gomock.Any()).Times(0)
	sub.EXPECT().OnComplete(gomock.Any()).Times(0)
	sub.EXPECT().OnError(gomock.Any(), gomock.Any()).Times(1)
	sub.EXPECT().OnSubscribe(gomock.Any(), gomock.Any()).Do(MockRequestInfinite).Times(1)

	empty := newMonoJust[int](1)
	predicate := func(ctx context.Context, next int) bool {
		panic("fake panic")
	}

	newMonoFilter[int](empty, predicate).SubscribeWith(context.TODO(), sub)
}
