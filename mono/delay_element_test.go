package mono

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/jjeffcaii/reactor-go"
	"github.com/jjeffcaii/reactor-go/scheduler"
	"github.com/stretchr/testify/assert"
)

func TestMonoDelayElement_SubscribeWith(t *testing.T) {
	done := make(chan struct{})
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	const value = 42

	s := NewMockSubscriber(ctrl)
	s.EXPECT().OnSubscribe(gomock.Any(), gomock.Any()).Do(MockRequestInfinite).Times(1)
	s.EXPECT().OnNext(gomock.Eq(value)).Times(1)
	s.EXPECT().OnComplete().Do(func() {
		done <- struct{}{}
	}).Times(1)
	s.EXPECT().OnError(gomock.Any()).Times(0)

	const delay = 10 * time.Millisecond
	startTime := time.Now()
	newMonoDelayElement(newMonoJust(value), delay, scheduler.Parallel()).SubscribeWith(context.Background(), s)

	<-done
	assert.Greater(t, time.Since(startTime).Nanoseconds(), delay.Nanoseconds())

	fakeErr := errors.New("fake error")
	s = NewMockSubscriber(ctrl)
	s.EXPECT().OnSubscribe(gomock.Any(), gomock.Any()).Do(MockRequestInfinite).Times(1)
	s.EXPECT().OnNext(gomock.Any()).Times(0)
	s.EXPECT().OnComplete().Times(0)
	s.EXPECT().OnError(gomock.Eq(fakeErr)).DoAndReturn(func(err error) error {
		close(done)
		return nil
	}).Times(1)

	startTime = time.Now()
	newMonoDelayElement(newMonoError(fakeErr), delay, scheduler.Parallel()).SubscribeWith(context.Background(), s)
	<-done
	assert.Less(t, time.Since(startTime).Nanoseconds(), delay.Nanoseconds())

	s = NewMockSubscriber(ctrl)
	s.EXPECT().OnSubscribe(gomock.Any(), gomock.Any()).Do(MockRequestInfinite).Times(1)
	s.EXPECT().OnNext(gomock.Any()).Times(0)
	s.EXPECT().OnComplete().Times(1)
	s.EXPECT().OnError(gomock.Any()).Times(0)

	newMonoDelayElement(newMonoEmpty(), delay, scheduler.Immediate()).SubscribeWith(context.Background(), s)
}

func TestMonoDelayElement_Cancel(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	onSubscribe := func(ctx context.Context, su reactor.Subscription) {
		su.Cancel()
	}

	s := NewMockSubscriber(ctrl)
	s.EXPECT().OnSubscribe(gomock.Any(), gomock.Any()).Do(onSubscribe).Times(1)
	s.EXPECT().OnNext(gomock.Any()).Times(0)
	s.EXPECT().OnComplete().Times(0)
	s.EXPECT().OnError(gomock.Eq(reactor.ErrSubscribeCancelled)).Times(1)

	startTime := time.Now()
	delay := 10 * time.Millisecond
	newMonoDelayElement(newMonoJust(1), delay, scheduler.Immediate()).SubscribeWith(context.Background(), s)
	assert.Less(t, time.Since(startTime).Nanoseconds(), delay.Nanoseconds())
}
