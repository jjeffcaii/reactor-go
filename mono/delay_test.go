package mono

import (
	"context"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/jjeffcaii/reactor-go"
	"github.com/stretchr/testify/assert"
)

func TestMonoDelay_SubscribeWith(t *testing.T) {
	done := make(chan struct{})
	onComplete := func() {
		close(done)
	}
	const delay = 10 * time.Millisecond

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	s := NewMockSubscriber(ctrl)
	s.EXPECT().OnError(gomock.Any()).Times(0)
	s.EXPECT().OnNext(gomock.Any()).Times(1)
	s.EXPECT().OnComplete().Do(onComplete).Times(1)
	s.EXPECT().OnSubscribe(gomock.Any(), gomock.Any()).Do(MockRequestInfinite).Times(1)

	startTime := time.Now()
	newMonoDelay(delay).SubscribeWith(context.Background(), s)
	<-done

	assert.Greater(t, time.Since(startTime).Nanoseconds(), delay.Nanoseconds())
}

func TestMonoDelay_Cancel(t *testing.T) {
	const delay = 10 * time.Millisecond
	onSubscribe := func(ctx context.Context, su reactor.Subscription) {
		su.Request(1)
		time.AfterFunc(delay/2, func() {
			su.Cancel()
		})
	}

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	s := NewMockSubscriber(ctrl)
	s.EXPECT().OnError(gomock.Eq(reactor.ErrSubscribeCancelled)).Times(1)
	s.EXPECT().OnNext(gomock.Any()).Times(0)
	s.EXPECT().OnComplete().Times(0)
	s.EXPECT().OnSubscribe(gomock.Any(), gomock.Any()).Do(onSubscribe).Times(1)

	newMonoDelay(delay).SubscribeWith(context.Background(), s)

	time.Sleep(2 * delay)
}
