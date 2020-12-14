package mono

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/jjeffcaii/reactor-go"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

func TestMonoPeek_SubscribeWith(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	s := NewMockSubscriber(ctrl)
	s.EXPECT().OnComplete().Times(1)
	s.EXPECT().OnSubscribe(gomock.Any(), gomock.Any()).Do(MockRequestInfinite).Times(1)
	s.EXPECT().OnNext(gomock.Any()).Times(1)
	s.EXPECT().OnError(gomock.Any()).Times(0)

	var (
		cntNext      int
		cntComplete  int
		cntSubscribe int
		cntRequest   int
		cntCancel    int
		cntError     int
	)

	newMonoPeek(newMonoJust(1), peekNext(func(v reactor.Any) error {
		assert.Equal(t, 1, v.(int), "bad value")
		cntNext++
		return nil
	}), peekComplete(func() {
		cntComplete++
	}), peekError(func(e error) {
		cntError++
	}), peekSubscribe(func(ctx context.Context, su reactor.Subscription) {
		cntSubscribe++
	}), peekRequest(func(n int) {
		cntRequest++
	}), peekCancel(func() {
		cntCancel++
	})).SubscribeWith(context.Background(), s)
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

	fakeErr := errors.New("fake error")

	s := NewMockSubscriber(ctrl)
	s.EXPECT().OnComplete().Times(0)
	s.EXPECT().OnSubscribe(gomock.Any(), gomock.Any()).Do(MockRequestInfinite).Times(1)
	s.EXPECT().OnNext(gomock.Any()).Times(0)
	s.EXPECT().OnError(gomock.Any()).Times(1)

	newMonoPeek(newMonoJust(1), peekNext(func(v reactor.Any) error {
		panic("fake panic")
	})).SubscribeWith(context.Background(), s)

	s = NewMockSubscriber(ctrl)
	s.EXPECT().OnComplete().Times(0)
	s.EXPECT().OnSubscribe(gomock.Any(), gomock.Any()).Do(MockRequestInfinite).Times(1)
	s.EXPECT().OnNext(gomock.Any()).Times(0)
	s.EXPECT().OnError(gomock.Any()).DoAndReturn(func(err error) error {
		assert.Equal(t, fakeErr, errors.Cause(err))
		return nil
	}).Times(1)

	newMonoPeek(newMonoJust(1), peekNext(func(v reactor.Any) error {
		panic(fakeErr)
	})).SubscribeWith(context.Background(), s)

	s = NewMockSubscriber(ctrl)
	s.EXPECT().OnComplete().Times(0)
	s.EXPECT().OnSubscribe(gomock.Any(), gomock.Any()).Do(MockRequestInfinite).Times(1)
	s.EXPECT().OnNext(gomock.Any()).Times(0)
	s.EXPECT().OnError(gomock.Any()).DoAndReturn(func(err error) error {
		assert.NotEqual(t, fakeErr, err)
		return nil
	}).Times(1)

	newMonoPeek(newMonoError(fakeErr), peekError(func(e error) {
		assert.Equal(t, fakeErr, e)
		panic("fake error")
	})).SubscribeWith(context.Background(), s)
}
