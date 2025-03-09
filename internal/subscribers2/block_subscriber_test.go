package subscribers

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/jjeffcaii/reactor-go"
	"github.com/jjeffcaii/reactor-go/internal"
)

var fakeErr = errors.New("fake error")

func TestBlockSubscriber(t *testing.T) {

	s := NewBlockSubscriber[int]()

	// complete test
	go func(ctx context.Context) {
		s.OnNext(ctx, 1)
		s.OnComplete(ctx)
	}(context.TODO())

	<-s.Done()

	assert.NoError(t, s.E, "should not return error")
	assert.True(t, s.Valid, "should be valid")
	assert.Equal(t, 1, s.V, "bad result")

	// error test
	s = NewBlockSubscriber[int]()
	s.OnError(context.TODO(), fakeErr)
	// omit
	s.OnNext(context.TODO(), 2)
	s.OnError(context.TODO(), fakeErr)
	s.OnComplete(context.TODO())

	<-s.Done()

	assert.Equal(t, fakeErr, s.E, "should be fake error")
	assert.False(t, s.Valid, "should be valid=false")

	// empty test
	s = NewBlockSubscriber[int]()
	s.OnComplete(context.TODO())
	// omit
	s.OnNext(context.TODO(), 2)
	s.OnError(context.TODO(), fakeErr)

	<-s.Done()

	assert.NoError(t, s.E, "should not return error")
	assert.False(t, s.Valid, "should be valid=false")
}

func TestBlockSubscriber_OnSubscribe(t *testing.T) {
	var su internal.EmptySubscription

	s := NewBlockSubscriber[int]()
	s.OnSubscribe(context.TODO(), su)

	s = NewBlockSubscriber[int]()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	s.OnSubscribe(ctx, su)
	<-s.Done()
	assert.Error(t, s.E, "should return error")
	assert.True(t, reactor.IsCancelledError(s.E), "should be cancelled error")

	s = NewBlockSubscriber[int]()
	ctx, cancel = context.WithCancel(context.Background())
	s.OnSubscribe(ctx, su)
	s.OnComplete(ctx)
	time.Sleep(10 * time.Millisecond)
	cancel()

	<-s.Done()
}
