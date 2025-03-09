package mono

import (
	"context"
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"

	"github.com/jjeffcaii/reactor-go"
	"github.com/jjeffcaii/reactor-go/testdata"
)

func TestMonoPeek_SubscribeWith(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	s := testdata.NewMockSubscriber[int](ctrl)
	s.EXPECT().OnComplete(gomock.Any()).Times(1)
	s.EXPECT().OnSubscribe(gomock.Any(), gomock.Any()).Do(MockRequestInfinite).Times(1)
	s.EXPECT().OnNext(gomock.Any(), gomock.Any()).Times(1)
	s.EXPECT().OnError(gomock.Any(), gomock.Any()).Times(0)

	var (
		cntNext      int
		cntComplete  int
		cntSubscribe int
		cntRequest   int
		cntCancel    int
		cntError     int
	)

	newMonoPeek[int](
		newMonoJust(1),
		peekNext[int](func(ctx context.Context, v int) error {
			assert.Equal(t, 1, v, "bad value")
			cntNext++
			return nil
		}),
		peekComplete[int](func(ctx context.Context) {
			cntComplete++
		}),
		peekError[int](func(ctx context.Context, e error) {
			cntError++
		}),
		peekSubscribe[int](func(ctx context.Context, su reactor.Subscription) {
			cntSubscribe++
		}),
		peekRequest[int](func(ctx context.Context, n int) {
			cntRequest++
		}),
		peekCancel[int](func(ctx context.Context) {
			cntCancel++
		}),
	).SubscribeWith(context.Background(), s)
	assert.Equal(t, 1, cntNext)
	assert.Equal(t, 1, cntComplete)
	assert.Equal(t, 1, cntSubscribe)
	assert.Equal(t, 1, cntRequest)
	assert.Zero(t, cntCancel)
	assert.Zero(t, cntError)
}

func TestPeek_Panic(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	s := testdata.NewMockSubscriber[int](ctrl)
	s.EXPECT().OnComplete(gomock.Any()).Times(0)
	s.EXPECT().OnSubscribe(gomock.Any(), gomock.Any()).Do(MockRequestInfinite).Times(1)
	s.EXPECT().OnNext(gomock.Any(), gomock.Any()).Times(0)
	s.EXPECT().OnError(gomock.Any(), gomock.Any()).Times(1)

	newMonoPeek[int](
		newMonoJust(1),
		peekNext[int](func(ctx context.Context, v int) error {
			panic("fake panic")
		})).
		SubscribeWith(context.Background(), s)

	s = testdata.NewMockSubscriber[int](ctrl)
	s.EXPECT().OnComplete(gomock.Any()).Times(0)
	s.EXPECT().OnSubscribe(gomock.Any(), gomock.Any()).Do(MockRequestInfinite).Times(1)
	s.EXPECT().OnNext(gomock.Any(), gomock.Any()).Times(0)
	s.EXPECT().
		OnError(gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, err error) error {
			assert.Equal(t, fakeErr, errors.Cause(err))
			return nil
		}).
		Times(1)

	newMonoPeek[int](
		newMonoJust(1),
		peekNext(func(ctx context.Context, v int) error {
			panic(fakeErr)
		})).
		SubscribeWith(context.Background(), s)

	s = testdata.NewMockSubscriber[int](ctrl)
	s.EXPECT().OnComplete(gomock.Any()).Times(0)
	s.EXPECT().OnSubscribe(gomock.Any(), gomock.Any()).Do(MockRequestInfinite).Times(1)
	s.EXPECT().OnNext(gomock.Any(), gomock.Any()).Times(0)
	s.EXPECT().
		OnError(gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, err error) error {
			assert.NotEqual(t, fakeErr, err)
			return nil
		}).
		Times(1)

	newMonoPeek[int](
		newMonoError[int](fakeErr),
		peekError[int](func(ctx context.Context, e error) {
			assert.Equal(t, fakeErr, e)
			panic("fake error")
		})).
		SubscribeWith(context.TODO(), s)
}
