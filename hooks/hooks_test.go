package hooks_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/jjeffcaii/reactor-go"
	"github.com/jjeffcaii/reactor-go/hooks"
	"github.com/jjeffcaii/reactor-go/mono"
	"github.com/stretchr/testify/assert"
)

type Any = reactor.Any

func TestCreate_OnNextDrop(t *testing.T) {
	var droppedValue int
	var droppedError error
	hooks.OnNextDrop(func(i Any) {
		droppedValue = i.(int)
	})
	hooks.OnErrorDrop(func(e error) {
		droppedError = e
	})

	done := make(chan struct{})

	mono.
		Create(func(context context.Context, sink mono.Sink) {
			sink.Success(1)
			// Should be dropped below
			sink.Success(2)
			sink.Error(errors.New("mock"))
			close(done)
		}).
		DoOnNext(func(v Any) error {
			assert.Equal(t, 1, v.(int), "bad doOnNext")
			return nil
		}).
		Subscribe(context.Background())
	<-done
	assert.Equal(t, 2, droppedValue, "bad onNextDrop")
	assert.Error(t, droppedError, "no OnErrorDrop happened")
}

func TestProcessor_OnNextDrop(t *testing.T) {
	done := make(chan struct{})
	var hasDrop bool
	hooks.OnNextDrop(func(i Any) {
		defer close(done)
		assert.Equal(t, 2, i.(int), "bad onNextDrop")
		hasDrop = true
	})
	p, s, d := mono.NewProcessor(nil, nil)
	defer d.Dispose()
	time.AfterFunc(100*time.Millisecond, func() {
		s.Success(1)
	})
	time.AfterFunc(200*time.Millisecond, func() {
		s.Success(2)
	})
	p.
		DoOnNext(func(v Any) error {
			assert.Equal(t, 1, v.(int), "bad doOnNext")
			return nil
		}).
		Subscribe(context.Background())
	<-done
	assert.True(t, hasDrop, "no doNextDrop happened")
}
