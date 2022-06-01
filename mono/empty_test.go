package mono

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
)

func TestMonoEmpty_SubscribeWith(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	s := NewMockSubscriber(ctrl)
	s.EXPECT().OnSubscribe(gomock.Any(), gomock.Any()).Do(MockRequestInfinite).Times(1)
	s.EXPECT().OnNext(gomock.Any()).Times(0)
	s.EXPECT().OnComplete().Times(1)
	s.EXPECT().OnError(gomock.Any()).Times(0)
	newMonoEmpty().SubscribeWith(context.Background(), s)
}
