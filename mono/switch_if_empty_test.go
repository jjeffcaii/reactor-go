package mono

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/jjeffcaii/reactor-go"
	"github.com/stretchr/testify/assert"
)

func TestSwitchIfEmpty_Empty(t *testing.T) {
	exec := func(t *testing.T, empty reactor.RawPublisher) {
		const fakeValue = 42
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		sub := NewMockSubscriber(ctrl)
		sub.EXPECT().OnSubscribe(gomock.Any(), gomock.Any()).Do(MockRequestInfinite).Times(1)
		sub.EXPECT().OnNext(gomock.Eq(fakeValue)).Times(1)
		sub.EXPECT().OnError(gomock.Any()).Times(0)
		sub.EXPECT().OnComplete().Times(1)

		newMonoSwitchIfEmpty(empty, newMonoJust(fakeValue)).SubscribeWith(context.Background(), sub)
	}

	exec(t, newMonoEmpty())
	exec(t, newMonoCreate(func(ctx context.Context, s Sink) {
		s.Success(nil)
	}))

	v, err := Empty().
		SwitchIfEmpty(Just(333)).
		Map(func(i Any) (Any, error) {
			return i.(int) * 2, nil
		}).
		Block(context.Background())
	assert.NoError(t, err, "should not return error")
	assert.Equal(t, 666, v.(int), "bad result")
}

func TestSwitchIfEmpty_NotEmpty(t *testing.T) {
	v, err := Just(666).SwitchIfEmpty(Just(444)).Block(context.Background())
	assert.NoError(t, err, "err occurred")
	assert.Equal(t, 666, v, "bad result")
}
