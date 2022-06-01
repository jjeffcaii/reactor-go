package subscribers

import (
	"context"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/jjeffcaii/reactor-go"
	"github.com/stretchr/testify/assert"
)

func TestDoFinallySubscriber(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	su := NewMockSubscription(ctrl)
	su.EXPECT().Request(gomock.Any()).Times(1)
	su.EXPECT().Cancel().Times(0)

	var prevOperation string

	actual := NewMockSubscriber(ctrl)
	actual.EXPECT().
		OnSubscribe(gomock.Any(), gomock.Any()).
		Do(func(ctx context.Context, su reactor.Subscription) {
			su.Request(reactor.RequestInfinite)
			prevOperation = "OnSubscribe"
		}).
		Times(1)
	actual.EXPECT().OnNext(gomock.Any()).Do(func(v reactor.Any) {
		prevOperation = "OnNext"
	}).Times(1)
	actual.EXPECT().OnError(gomock.Any()).Times(0)
	actual.EXPECT().OnComplete().Do(func() {
		prevOperation = "OnComplete"
	}).Times(1)

	s := NewDoFinallySubscriber(actual, func(signal reactor.SignalType) {
		prevOperation = "DoFinally"
		assert.Equal(t, reactor.SignalTypeComplete, signal)
	})
	s.OnSubscribe(context.Background(), su)
	s.OnNext(42)
	s.OnComplete()

	assert.Equal(t, "DoFinally", prevOperation)

	prevOperation = ""
	su = NewMockSubscription(ctrl)
	su.EXPECT().Request(gomock.Any()).Times(1)
	su.EXPECT().Cancel().Times(0)
	actual = NewMockSubscriber(ctrl)
	actual.EXPECT().
		OnSubscribe(gomock.Any(), gomock.Any()).
		Do(func(ctx context.Context, su reactor.Subscription) {
			su.Request(reactor.RequestInfinite)
			prevOperation = "OnSubscribe"
		}).
		Times(1)
	actual.EXPECT().OnNext(gomock.Any()).Do(func(v reactor.Any) {
		prevOperation = "OnNext"
	}).Times(0)
	actual.EXPECT().OnError(gomock.Any()).Do(func(e error) {
		prevOperation = "OnError"
	}).Times(1)
	actual.EXPECT().OnComplete().Do(func() {
		prevOperation = "OnComplete"
	}).Times(0)
	s = NewDoFinallySubscriber(actual, func(signal reactor.SignalType) {
		prevOperation = "DoFinally"
		assert.Equal(t, reactor.SignalTypeError, signal)
	})

	fakeErr := errors.New("fake error")
	s.OnSubscribe(context.Background(), su)
	s.OnError(fakeErr)

	assert.Equal(t, "DoFinally", prevOperation)
}

func TestDoFinallySubscriberPool_Put(t *testing.T) {
	assert.NotPanics(t, func() {
		globalDoFinallySubscriberPool.put(nil)
	})
}

func TestDoFinallySubscriber_Cancel(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	su := NewMockSubscription(ctrl)
	su.EXPECT().Request(gomock.Any()).Times(0)
	su.EXPECT().Cancel().Times(1)

	actual := NewMockSubscriber(ctrl)
	actual.EXPECT().OnSubscribe(gomock.Any(), gomock.Any()).Do(func(ctx context.Context, su reactor.Subscription) {
		su.Cancel()
	}).Times(1)
	actual.EXPECT().OnError(gomock.Any()).AnyTimes()
	actual.EXPECT().OnNext(gomock.Any()).Times(0)
	actual.EXPECT().OnComplete().Times(0)

	var doFinallyCalls int

	s := NewDoFinallySubscriber(actual, func(signal reactor.SignalType) {
		assert.Equal(t, reactor.SignalTypeCancel, signal)
		doFinallyCalls++
	})
	s.Cancel()
	s.OnSubscribe(context.Background(), su)

	assert.Equal(t, 1, doFinallyCalls)
}
