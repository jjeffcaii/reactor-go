package subscribers

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/jjeffcaii/reactor-go"
	"github.com/jjeffcaii/reactor-go/internal"
	"github.com/stretchr/testify/assert"
)

func TestBlockSubscriber(t *testing.T) {
	fakeErr := errors.New("fake error")

	// complete test
	s := BorrowBlockSubscriber()
	go func() {
		s.OnNext(1)
		s.OnComplete()
	}()

	<-s.Done()

	assert.NoError(t, s.E, "should not return error")
	assert.Equal(t, 1, s.V, "bad result")
	ReturnBlockSubscriber(s)

	// error test
	s = BorrowBlockSubscriber()
	s.OnError(fakeErr)
	// omit
	s.OnNext(2)
	s.OnError(fakeErr)
	s.OnComplete()

	<-s.Done()

	assert.Equal(t, fakeErr, s.E, "should be fake error")
	assert.Nil(t, s.V)
	ReturnBlockSubscriber(s)

	// empty test
	s = BorrowBlockSubscriber()
	s.OnComplete()
	// omit
	s.OnNext(2)
	s.OnError(fakeErr)

	<-s.Done()

	assert.NoError(t, s.E, "should not return error")
	assert.Nil(t, s.V)
	ReturnBlockSubscriber(s)
}

func TestReturnBlockSubscriber(t *testing.T) {
	assert.NotPanics(t, func() {
		ReturnBlockSubscriber(nil)
	})
}

func TestBlockSubscriber_OnSubscribe(t *testing.T) {
	s := BorrowBlockSubscriber()
	s.OnSubscribe(context.Background(), internal.EmptySubscription)
	ReturnBlockSubscriber(s)

	s = BorrowBlockSubscriber()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	s.OnSubscribe(ctx, internal.EmptySubscription)
	<-s.Done()
	assert.Error(t, s.E, "should return error")
	assert.True(t, reactor.IsCancelledError(s.E), "should be cancelled error")
	ReturnBlockSubscriber(s)

	s = BorrowBlockSubscriber()
	ctx, cancel = context.WithCancel(context.Background())
	s.OnSubscribe(ctx, internal.EmptySubscription)
	s.OnComplete()
	time.Sleep(10 * time.Millisecond)
	cancel()

	<-s.Done()
}
