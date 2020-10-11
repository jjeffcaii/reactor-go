package mono_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/jjeffcaii/reactor-go"
	"github.com/jjeffcaii/reactor-go/mono"
	"github.com/stretchr/testify/assert"
)

func TestProcessor(t *testing.T) {
	p := mono.CreateProcessor()

	time.AfterFunc(100*time.Millisecond, func() {
		p.Success(333)
	})

	v, err := p.
		Map(func(i Any) (Any, error) {
			return i.(int) * 2, nil
		}).
		Block(context.Background())
	assert.NoError(t, err, "block failed")
	assert.Equal(t, 666, v.(int), "bad result")

	assert.Panics(t, func() {
		p.Subscribe(context.Background())
	})
}

func TestProcessor_Context(t *testing.T) {
	p := mono.CreateProcessor()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	done := make(chan struct{})
	p.
		DoOnError(func(e error) {
			assert.True(t, reactor.IsCancelledError(e))
		}).
		DoFinally(func(signal reactor.SignalType) {
			assert.Equal(t, reactor.SignalTypeCancel, signal)
			close(done)
		}).
		Subscribe(ctx)
	<-done
}

func TestProcessor_Error(t *testing.T) {
	fakeErr := errors.New("fake error")
	p := mono.CreateProcessor()
	done := make(chan error, 1)
	p.
		DoOnError(func(e error) {
			done <- e
		}).
		Subscribe(context.Background())

	time.Sleep(100 * time.Millisecond)
	p.Error(fakeErr)
	e := <-done
	assert.Equal(t, fakeErr, e)
}
