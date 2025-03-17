package mono_test

import (
	"context"
	"io"
	"testing"
	"time"

	"go.uber.org/mock/gomock"

	mono "github.com/jjeffcaii/reactor-go/mono2"
	"github.com/jjeffcaii/reactor-go/testdata"
)

func TestMonoTimeout_SubscribeWith(t *testing.T) {
	source := mono.Create[int](func(ctx context.Context, sink mono.Sink[int]) {
		time.Sleep(15 * time.Millisecond)
		sink.Success(1)
	})

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	s := testdata.NewMockSubscriber[int](ctrl)
	s.EXPECT().OnSubscribe(gomock.Any(), gomock.Any()).Do(testdata.MockRequestInfinite).Times(1)
	s.EXPECT().OnError(gomock.Any(), gomock.Any()).Times(1)
	s.EXPECT().OnNext(gomock.Any(), gomock.Any()).Times(0)
	s.EXPECT().OnComplete(gomock.Any()).Times(0)

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
	t.Run("ErrorCompleteNext", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		ms := NewMockSubscription(ctrl)

		ms.EXPECT().Request(gomock.Any()).AnyTimes()
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
	})

	t.Run("CompleteErrorNext", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		ms := NewMockSubscription(ctrl)

		ms.EXPECT().Request(gomock.Any()).AnyTimes()
		ms.EXPECT().Cancel().AnyTimes()

		msub := NewMockSubscriber(ctrl)
		msub.EXPECT().OnSubscribe(gomock.Any(), gomock.Any()).Do(MockRequestInfinite).Times(1)
		msub.EXPECT().OnError(gomock.Any()).Times(0)
		msub.EXPECT().OnNext(gomock.Any()).Times(0)
		msub.EXPECT().OnComplete().Times(1)

		sub := &timeoutSubscriber{
			actual: msub,
			done:   make(chan struct{}),
		}

		sub.OnSubscribe(context.Background(), ms)

		sub.OnComplete()
		sub.OnError(io.EOF)
		sub.OnNext(1)
	})

	t.Run("NextErrorComplete", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		ms := NewMockSubscription(ctrl)

		ms.EXPECT().Request(gomock.Any()).AnyTimes()
		ms.EXPECT().Cancel().AnyTimes()

		msub := NewMockSubscriber(ctrl)
		msub.EXPECT().OnSubscribe(gomock.Any(), gomock.Any()).Do(MockRequestInfinite).Times(1)
		msub.EXPECT().OnError(gomock.Any()).Times(0)
		msub.EXPECT().OnNext(gomock.Any()).Times(1)
		msub.EXPECT().OnComplete().Times(1)

		sub := &timeoutSubscriber{
			actual: msub,
			done:   make(chan struct{}),
		}

		sub.OnSubscribe(context.Background(), ms)

		sub.OnNext(1)
		sub.OnError(io.EOF)
		sub.OnComplete()
	})

}
