package subscribers

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"

	"github.com/jjeffcaii/reactor-go"
	"github.com/jjeffcaii/reactor-go/testdata"
)

func TestDoFinallySubscriber(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	su := testdata.NewMockSubscription(ctrl)
	su.EXPECT().Request(gomock.Any(), gomock.Any()).Times(1)
	su.EXPECT().Cancel(gomock.Any()).Times(0)

	var prevOperation string

	actual := testdata.NewMockSubscriber[int](ctrl)
	actual.EXPECT().
		OnSubscribe(gomock.Any(), gomock.Any()).
		Do(func(ctx context.Context, su reactor.Subscription) {
			su.Request(ctx, reactor.RequestInfinite)
			prevOperation = "OnSubscribe"
		}).
		Times(1)
	actual.EXPECT().
		OnNext(gomock.Any(), gomock.Any()).
		Do(func(ctx context.Context, v int) error {
			prevOperation = "OnNext"
			return nil
		}).
		Times(1)
	actual.EXPECT().OnError(gomock.Any(), gomock.Any()).Times(0)
	actual.EXPECT().
		OnComplete(gomock.Any()).
		Do(func(ctx context.Context) {
			prevOperation = "OnComplete"
		}).
		Times(1)

	s := NewDoFinallySubscriber[int](actual, func(ctx context.Context, signal reactor.SignalType) {
		prevOperation = "DoFinally"
		assert.Equal(t, reactor.SignalTypeComplete, signal)
	})
	s.OnSubscribe(context.TODO(), su)
	s.OnNext(context.TODO(), 42)
	s.OnComplete(context.TODO())

	assert.Equal(t, "DoFinally", prevOperation)

	prevOperation = ""
	su = testdata.NewMockSubscription(ctrl)
	su.EXPECT().Request(gomock.Any(), gomock.Any()).Times(1)
	su.EXPECT().Cancel(gomock.Any()).Times(0)
	actual = testdata.NewMockSubscriber[int](ctrl)
	actual.EXPECT().
		OnSubscribe(gomock.Any(), gomock.Any()).
		Do(func(ctx context.Context, su reactor.Subscription) {
			su.Request(ctx, reactor.RequestInfinite)
			prevOperation = "OnSubscribe"
		}).
		Times(1)
	actual.EXPECT().
		OnNext(gomock.Any(), gomock.Any()).
		Do(func(ctx context.Context, v int) error {
			prevOperation = "OnNext"
			return nil
		}).
		Times(0)
	actual.EXPECT().
		OnError(gomock.Any(), gomock.Any()).
		Do(func(ctx context.Context, e error) {
			prevOperation = "OnError"
		}).
		Times(1)
	actual.EXPECT().
		OnComplete(gomock.Any()).
		Do(func(ctx context.Context) {
			prevOperation = "OnComplete"
		}).
		Times(0)
	s = NewDoFinallySubscriber[int](actual, func(ctx context.Context, signal reactor.SignalType) {
		prevOperation = "DoFinally"
		assert.Equal(t, reactor.SignalTypeError, signal)
	})

	s.OnSubscribe(context.TODO(), su)
	s.OnError(context.TODO(), fakeErr)

	assert.Equal(t, "DoFinally", prevOperation)
}

func TestDoFinallySubscriber_Cancel(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	su := testdata.NewMockSubscription(ctrl)
	su.EXPECT().Request(gomock.Any(), gomock.Any()).Times(0)
	su.EXPECT().Cancel(gomock.Any()).Times(1)

	actual := testdata.NewMockSubscriber[int](ctrl)
	actual.EXPECT().
		OnSubscribe(gomock.Any(), gomock.Any()).
		Do(func(ctx context.Context, su reactor.Subscription) {
			su.Cancel(ctx)
		}).
		Times(1)
	actual.EXPECT().OnError(gomock.Any(), gomock.Any()).AnyTimes()
	actual.EXPECT().OnNext(gomock.Any(), gomock.Any()).Times(0)
	actual.EXPECT().OnComplete(gomock.Any()).Times(0)

	var doFinallyCalls int

	s := NewDoFinallySubscriber[int](actual, func(ctx context.Context, signal reactor.SignalType) {
		assert.Equal(t, reactor.SignalTypeCancel, signal)
		doFinallyCalls++
	})
	s.Cancel(context.TODO())
	s.OnSubscribe(context.TODO(), su)

	assert.Equal(t, 1, doFinallyCalls)
}
