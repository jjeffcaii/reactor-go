package mono

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"

	"github.com/jjeffcaii/reactor-go"
	"github.com/jjeffcaii/reactor-go/testdata"
)

func TestDoFinally(t *testing.T) {
	const fakeValue = 42
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	sub := testdata.NewMockSubscriber[int](ctrl)
	sub.EXPECT().OnSubscribe(gomock.Any(), gomock.Any()).Do(MockRequestInfinite).Times(1)
	sub.EXPECT().OnNext(gomock.Any(), gomock.Any()).Times(1)
	sub.EXPECT().OnError(gomock.Any(), gomock.Any()).Times(0)
	sub.EXPECT().OnComplete(gomock.Any()).Times(1)

	var seq []int
	Just(fakeValue).
		DoOnNext(func(ctx context.Context, v int) error {
			seq = append(seq, 1)
			return nil
		}).
		DoFinally(func(ctx context.Context, s reactor.SignalType) {
			seq = append(seq, 3)
		}).
		DoOnNext(func(ctx context.Context, v int) error {
			seq = append(seq, 2)
			return nil
		}).
		SubscribeWith(context.TODO(), sub)

	assert.Equal(t, []int{1, 2, 3}, seq, "wrong execute order")
}
