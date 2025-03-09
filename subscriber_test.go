package reactor

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

var fakeErr = errors.New("fake error")

type fakeSubscription struct {
	r, c int
}

func (f *fakeSubscription) Request(ctx context.Context, n int) {
	f.r += n
}

func (f *fakeSubscription) Cancel(ctx context.Context) {
	f.c++
}

func TestSubscriber(t *testing.T) {
	s := NewSubscriber[int]()
	assert.NotPanics(t, func() {
		s.OnNext(context.TODO(), 1)
		s.OnComplete(context.TODO())
		s.OnError(context.TODO(), fakeErr)
		s.OnSubscribe(context.Background(), &fakeSubscription{})
	})

	var cntOnNext, cntOnSubscribe, cntOnError, cntOnComplete int

	fsu := &fakeSubscription{}
	s = NewSubscriber[int](
		OnNext[int](func(ctx context.Context, v int) error {
			cntOnNext++
			assert.Equal(t, 1, v)
			return nil
		}),
		OnSubscribe[int](func(ctx context.Context, su Subscription) {
			cntOnSubscribe++
			assert.Equal(t, fsu, su)
		}),
		OnError[int](func(ctx context.Context, e error) {
			cntOnError++
			assert.Equal(t, fakeErr, e)
		}),
		OnComplete[int](func(ctx context.Context) {
			cntOnComplete++
		}),
	)
	s.OnNext(context.TODO(), 1)
	s.OnComplete(context.TODO())
	s.OnError(context.TODO(), fakeErr)
	s.OnSubscribe(context.TODO(), fsu)
	for _, calls := range []int{cntOnNext, cntOnSubscribe, cntOnError, cntOnComplete} {
		assert.Equal(t, 1, calls)
	}
}

func TestSubscriber_Nil(t *testing.T) {
	var s subscriber[int]
	assert.NotPanics(t, func() {
		s.OnComplete(context.TODO())
		s.OnError(context.TODO(), fakeErr)
		s.OnNext(context.TODO(), 1)
		s.OnSubscribe(context.Background(), &fakeSubscription{})
	})
}

func TestSubscriber_OnNextReturnsError(t *testing.T) {
	var cntOnError int
	s := NewSubscriber[int](
		OnNext(func(ctx context.Context, v int) error {
			return fakeErr
		}),
		OnError[int](func(ctx context.Context, e error) {
			assert.Equal(t, fakeErr, e, "should be fake error")
			cntOnError++
		}))
	s.OnNext(context.TODO(), 1)
	assert.Equal(t, 1, cntOnError)
}
