package mono

import (
	"context"
	"testing"

	"go.uber.org/mock/gomock"

	"github.com/jjeffcaii/reactor-go/testdata"
)

func TestMonoEmpty_SubscribeWith(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	s := testdata.NewMockSubscriber[int](ctrl)
	s.EXPECT().OnSubscribe(gomock.Any(), gomock.Any()).Do(MockRequestInfinite).Times(1)
	s.EXPECT().OnNext(gomock.Any(), gomock.Any()).Times(0)
	s.EXPECT().OnComplete(gomock.Any()).Times(1)
	s.EXPECT().OnError(gomock.Any(), gomock.Any()).Times(0)
	
	newMonoEmpty[int]().SubscribeWith(context.TODO(), s)
}
