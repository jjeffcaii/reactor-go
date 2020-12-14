package mono

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/jjeffcaii/reactor-go"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

func TestMonoMap_SubscribeWith(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	const value = 42
	tf := func(any reactor.Any) (reactor.Any, error) {
		return any.(int) * 2, nil
	}

	// happy path with value
	s := NewMockSubscriber(ctrl)
	s.EXPECT().OnSubscribe(gomock.Any(), gomock.Any()).Do(MockRequestInfinite).Times(1)
	s.EXPECT().OnError(gomock.Any()).Times(0)
	s.EXPECT().OnNext(gomock.Eq(value * 2)).Times(1)
	s.EXPECT().OnComplete().Times(1)

	newMonoMap(newMonoJust(value), tf).SubscribeWith(context.Background(), s)

	// happy path with error
	fakeErr := errors.New("fake error")
	s = NewMockSubscriber(ctrl)
	s.EXPECT().OnSubscribe(gomock.Any(), gomock.Any()).Do(MockRequestInfinite).Times(1)
	s.EXPECT().OnError(gomock.Eq(fakeErr)).Times(1)
	s.EXPECT().OnNext(gomock.Any()).Times(0)
	s.EXPECT().OnComplete().Times(0)

	newMonoMap(newMonoError(fakeErr), tf).SubscribeWith(context.Background(), s)
}

func TestMonoMap_Panic(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	s := NewMockSubscriber(ctrl)
	s.EXPECT().OnSubscribe(gomock.Any(), gomock.Any()).Do(MockRequestInfinite).Times(1)
	s.EXPECT().OnError(gomock.Any()).Times(1)
	s.EXPECT().OnNext(gomock.Any()).Times(0)
	s.EXPECT().OnComplete().Times(0)

	newMonoMap(newMonoJust(1), func(any reactor.Any) (reactor.Any, error) {
		panic("fake panic")
	}).SubscribeWith(context.Background(), s)

	fakeErr := errors.New("fake error")
	s = NewMockSubscriber(ctrl)
	s.EXPECT().OnSubscribe(gomock.Any(), gomock.Any()).Do(MockRequestInfinite).Times(1)
	s.EXPECT().OnError(gomock.Any()).DoAndReturn(func(err error) error {
		assert.Equal(t, fakeErr, errors.Cause(err))
		return nil
	}).Times(1)
	s.EXPECT().OnNext(gomock.Any()).Times(0)
	s.EXPECT().OnComplete().Times(0)

	newMonoMap(newMonoJust(1), func(any reactor.Any) (reactor.Any, error) {
		panic(fakeErr)
	}).SubscribeWith(context.Background(), s)
}

func TestMonoMapSubscriberPool_PutWithNilValue(t *testing.T) {
	assert.NotPanics(t, func() {
		globalMapSubscriberPool.put(nil)
	})
}

func TestMonoMap_Cancel(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	onSubscribe := func(ctx context.Context, su reactor.Subscription) {
		su.Cancel()
	}

	s := NewMockSubscriber(ctrl)
	s.EXPECT().OnSubscribe(gomock.Any(), gomock.Any()).Do(onSubscribe).Times(1)
	s.EXPECT().OnError(gomock.Eq(reactor.ErrSubscribeCancelled)).Times(1)
	s.EXPECT().OnNext(gomock.Any()).Times(0)
	s.EXPECT().OnComplete().Times(0)

	newMonoMap(newMonoJust(1), func(any reactor.Any) (reactor.Any, error) {
		return any.(int) * 2, nil
	}).SubscribeWith(context.Background(), s)
}
