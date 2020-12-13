package mono_test

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/jjeffcaii/reactor-go"
	"github.com/jjeffcaii/reactor-go/mono"
	"github.com/stretchr/testify/assert"
)

func TestDoFinally(t *testing.T) {
	const fakeValue = 42
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	sub := mono.NewMockSubscriber(ctrl)
	sub.EXPECT().OnSubscribe(gomock.Any(), gomock.Any()).Do(mono.MockRequestInfinite).Times(1)
	sub.EXPECT().OnNext(gomock.Any()).Times(1)
	sub.EXPECT().OnError(gomock.Any()).Times(0)
	sub.EXPECT().OnComplete().Times(1)

	var seq []int
	mono.Just(fakeValue).
		DoOnNext(func(v reactor.Any) error {
			seq = append(seq, 1)
			return nil
		}).
		DoFinally(func(s reactor.SignalType) {
			seq = append(seq, 3)
		}).
		DoOnNext(func(v reactor.Any) error {
			seq = append(seq, 2)
			return nil
		}).
		SubscribeWith(context.Background(), sub)
	assert.Equal(t, []int{1, 2, 3}, seq, "wrong execute order")
}
