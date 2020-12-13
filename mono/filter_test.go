package mono

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/jjeffcaii/reactor-go"
)

func TestMonoFilter(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	sub := NewMockSubscriber(ctrl)

	sub.EXPECT().OnNext(gomock.Eq(42)).Times(1)
	sub.EXPECT().OnComplete().Times(1)
	sub.EXPECT().OnError(gomock.Any()).Times(0)
	sub.EXPECT().OnSubscribe(gomock.Any(), gomock.Any()).Do(MockRequestInfinite).Times(1)

	predicate := func(any reactor.Any) bool {
		return any.(int) > 0
	}
	newMonoFilter(newMonoJust(42), predicate).SubscribeWith(context.Background(), sub)

	sub = NewMockSubscriber(ctrl)
	sub.EXPECT().OnNext(gomock.Any()).Times(0)
	sub.EXPECT().OnComplete().Times(1)
	sub.EXPECT().OnError(gomock.Any()).Times(0)
	sub.EXPECT().OnSubscribe(gomock.Any(), gomock.Any()).Do(MockRequestInfinite).Times(1)

	newMonoFilter(newMonoJust(-42), predicate).SubscribeWith(context.Background(), sub)
}
