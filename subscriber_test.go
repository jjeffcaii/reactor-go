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

func (f *fakeSubscription) Request(n int) {
	f.r += n
}

func (f *fakeSubscription) Cancel() {
	f.c++
}

func TestSubscriber(t *testing.T) {
	s := NewSubscriber()
	assert.NotPanics(t, func() {
		s.OnNext(1)
		s.OnNext(nil)
		s.OnComplete()
		s.OnError(fakeErr)
		s.OnSubscribe(context.Background(), &fakeSubscription{})
	})

	var cntOnNext, cntOnSubscribe, cntOnError, cntOnComplete int

	fsu := &fakeSubscription{}
	s = NewSubscriber(
		OnNext(func(v Any) error {
			cntOnNext++
			assert.Equal(t, 1, v)
			return nil
		}),
		OnSubscribe(func(ctx context.Context, su Subscription) {
			cntOnSubscribe++
			assert.Equal(t, fsu, su)
		}),
		OnError(func(e error) {
			cntOnError++
			assert.Equal(t, fakeErr, e)
		}),
		OnComplete(func() {
			cntOnComplete++
		}),
	)
	s.OnNext(1)
	s.OnComplete()
	s.OnError(fakeErr)
	s.OnSubscribe(context.Background(), fsu)
	for _, calls := range []int{cntOnNext, cntOnSubscribe, cntOnError, cntOnComplete} {
		assert.Equal(t, 1, calls)
	}
}

func TestSubscriber_Nil(t *testing.T) {
	s := (*subscriber)(nil)
	assert.NotPanics(t, func() {
		s.OnComplete()
		s.OnError(fakeErr)
		s.OnNext(1)
		s.OnSubscribe(context.Background(), &fakeSubscription{})
	})
}

func TestSubscriber_OnNextReturnsError(t *testing.T) {
	var cntOnError int
	s := NewSubscriber(OnNext(func(v Any) error {
		return fakeErr
	}), OnError(func(e error) {
		assert.Equal(t, fakeErr, e, "should be fake error")
		cntOnError++
	}))
	s.OnNext(1)
	assert.Equal(t, 1, cntOnError)
}
