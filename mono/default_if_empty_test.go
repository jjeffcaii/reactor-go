package mono

import (
	"context"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/jjeffcaii/reactor-go"
)

func TestMonoDefaultIfEmpty_SubscribeWith(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	var s *MockSubscriber

	s = NewMockSubscriber(ctrl)
	s.EXPECT().OnSubscribe(gomock.Any(), gomock.Any()).Do(MockRequestInfinite).Times(1)
	s.EXPECT().OnNext(gomock.Eq(333)).Times(1)
	s.EXPECT().OnError(gomock.Any()).Times(0)
	s.EXPECT().OnComplete().Times(1)
	newMonoDefaultIfEmpty(newMonoEmpty(), 333).SubscribeWith(context.Background(), s)

	s = NewMockSubscriber(ctrl)
	s.EXPECT().OnSubscribe(gomock.Any(), gomock.Any()).Do(MockRequestInfinite).Times(1)
	s.EXPECT().OnNext(gomock.Eq(333)).Times(1)
	s.EXPECT().OnError(gomock.Any()).Times(0)
	s.EXPECT().OnComplete().Times(1)
	newMonoDefaultIfEmpty(
		newMonoFilter(
			newMonoJust(-1),
			func(any reactor.Any) bool {
				return any.(int) >= 0
			},
		),
		333,
	).SubscribeWith(context.Background(), s)

	s = NewMockSubscriber(ctrl)
	s.EXPECT().OnSubscribe(gomock.Any(), gomock.Any()).Do(MockRequestInfinite).Times(1)
	s.EXPECT().OnNext(gomock.Eq(222)).Times(1)
	s.EXPECT().OnError(gomock.Any()).Times(0)
	s.EXPECT().OnComplete().Times(1)
	newMonoDefaultIfEmpty(newMonoJust(222), 333).SubscribeWith(context.Background(), s)

	fakeErr := errors.New("fake error")
	s = NewMockSubscriber(ctrl)
	s.EXPECT().OnSubscribe(gomock.Any(), gomock.Any()).Do(MockRequestInfinite).Times(1)
	s.EXPECT().OnNext(gomock.Any()).Times(0)
	s.EXPECT().OnError(gomock.Eq(fakeErr)).Times(1)
	s.EXPECT().OnComplete().Times(0)
	newMonoDefaultIfEmpty(newMonoError(fakeErr), 333).SubscribeWith(context.Background(), s)
}
