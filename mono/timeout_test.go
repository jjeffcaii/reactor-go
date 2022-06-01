package mono

import (
	"context"
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
