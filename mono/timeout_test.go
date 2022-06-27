package mono

import (
	"context"
	"io"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
)

func TestMonoTimeout_SubscribeWith(t *testing.T) {
	source := newMonoCreate(func(ctx context.Context, s Sink) {
		time.Sleep(15 * time.Millisecond)
		s.Success(1)
	})

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	s := NewMockSubscriber(ctrl)
	s.EXPECT().OnSubscribe(gomock.Any(), gomock.Any()).Do(MockRequestInfinite).Times(1)
	s.EXPECT().OnError(gomock.Any()).Times(1)
	s.EXPECT().OnNext(gomock.Any()).Times(0)
	s.EXPECT().OnComplete().Times(0)

	newMonoTimeout(source, 10*time.Millisecond).SubscribeWith(context.Background(), s)
	time.Sleep(20 * time.Millisecond)

	s = NewMockSubscriber(ctrl)
	s.EXPECT().OnSubscribe(gomock.Any(), gomock.Any()).Do(MockRequestInfinite).Times(1)
	s.EXPECT().OnError(gomock.Any()).Times(0)
	s.EXPECT().OnNext(gomock.Any()).Times(1)
	s.EXPECT().OnComplete().Times(1)
	newMonoTimeout(source, 20*time.Millisecond).SubscribeWith(context.Background(), s)
	time.Sleep(30 * time.Millisecond)
}

func TestTimeoutSubscriber(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ms := NewMockSubscription(ctrl)

	ms.EXPECT().Request(gomock.Any()).Times(1)
	ms.EXPECT().Cancel().AnyTimes()

	msub := NewMockSubscriber(ctrl)
	msub.EXPECT().OnSubscribe(gomock.Any(), gomock.Any()).Do(MockRequestInfinite).Times(1)
	msub.EXPECT().OnError(gomock.Any()).Times(1)
	msub.EXPECT().OnNext(gomock.Any()).Times(0)
	msub.EXPECT().OnComplete().Times(0)

	sub := &timeoutSubscriber{
		actual: msub,
		done:   make(chan struct{}),
	}

	sub.OnSubscribe(context.Background(), ms)

	sub.OnError(io.EOF)
	sub.OnComplete()
	sub.OnNext(1)
}
