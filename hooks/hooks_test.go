package hooks_test

import (
	"context"
	"testing"
	"time"

	"github.com/jjeffcaii/reactor-go/hooks"
	"github.com/jjeffcaii/reactor-go/mono"
	"github.com/stretchr/testify/assert"
)

func TestCreate_OnNextDrop(t *testing.T) {
	var hasDrop bool
	hooks.OnNextDrop(func(i interface{}) {
		assert.Equal(t, 2, i.(int), "bad onNextDrop")
		hasDrop = true
	})
	mono.
		Create(func(context context.Context, sink mono.Sink) {
			sink.Success(1)
			sink.Success(2)
		}).
		DoOnNext(func(v interface{}) {
			assert.Equal(t, 1, v.(int), "bad doOnNext")
		}).
		Subscribe(context.Background())
	assert.True(t, hasDrop, "no onNextDrop happened")
}

func TestProcessor_OnNextDrop(t *testing.T) {
	done := make(chan struct{})
	var hasDrop bool
	hooks.OnNextDrop(func(i interface{}) {
		defer close(done)
		assert.Equal(t, 2, i.(int), "bad onNextDrop")
		hasDrop = true
	})
	p := mono.CreateProcessor()
	time.AfterFunc(100*time.Millisecond, func() {
		p.Success(1)
	})
	time.AfterFunc(200*time.Millisecond, func() {
		p.Success(2)
	})
	p.
		DoOnNext(func(v interface{}) {
			assert.Equal(t, 1, v.(int), "bad doOnNext")
		}).
		Subscribe(context.Background())
	<-done
	assert.True(t, hasDrop, "no doNextDrop happened")
}
