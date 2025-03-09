package mono

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"

	"github.com/jjeffcaii/reactor-go"
	"github.com/jjeffcaii/reactor-go/testdata"
)

func TestSwitchIfEmpty_Empty(t *testing.T) {
	exec := func(t *testing.T, empty reactor.RawPublisher[int]) {
		const fakeValue = 42
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		testdata.NewMockSubscriber[int](ctrl).EXPECT().OnSubscribe(gomock.Any(), gomock.Any()).Do(MockRequestInfinite).Times(1)
		testdata.NewMockSubscriber[int](ctrl).EXPECT().OnNext(gomock.Any(), gomock.Eq(fakeValue)).Times(1)
		testdata.NewMockSubscriber[int](ctrl).EXPECT().OnError(gomock.Any(), gomock.Any()).Times(0)
		testdata.NewMockSubscriber[int](ctrl).EXPECT().OnComplete(gomock.Any()).Times(1)

		newMonoSwitchIfEmpty(empty, newMonoJust(fakeValue)).SubscribeWith(context.Background(), testdata.NewMockSubscriber[int](ctrl))
	}

	exec(t, newMonoEmpty[int]())
	exec(t, newMonoCreate[int](func(ctx context.Context, s Sink[int]) {
		s.Error(nil)
	}))

	v, err := Empty[int]().
		SwitchIfEmpty(Just(333)).
		Map(func(ctx context.Context, next int) (int, error) {
			return next * 2, nil
		}).
		Block(context.Background())
	assert.NoError(t, err, "should not return error")
	assert.True(t, v.Valid, "ok should be true")
	assert.Equal(t, 666, v.V, "bad result")
}

func TestSwitchIfEmpty_NotEmpty(t *testing.T) {
	v, err := Just(666).SwitchIfEmpty(Just(444)).Block(context.Background())
	assert.NoError(t, err, "err occurred")
	assert.Equal(t, 666, v.V, "bad result")
}
