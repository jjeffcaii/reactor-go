package flux_test

import (
	"context"
	"testing"
	"time"

	"github.com/jjeffcaii/reactor-go/flux"
	"github.com/stretchr/testify/assert"
)

func TestTake(t *testing.T) {
	var amount int
	var completed bool
	var cancelled bool
	_, err := flux.Interval(10 * time.Millisecond).
		Take(3).
		DoOnCancel(func() {
			cancelled = true
		}).
		DoOnComplete(func() {
			completed = true
		}).
		DoOnNext(func(v Any) error {
			amount++
			return nil
		}).
		BlockLast(context.Background())
	assert.NoError(t, err, "block last failed")
	assert.Equal(t, 3, amount, "bad amount ")
	assert.True(t, completed, "bad completed")
	assert.False(t, cancelled, "bad cancelled")
}

func TestTakeFromError(t *testing.T) {
	_, err := flux.Error(fakeErr).Take(3).BlockLast(context.Background())
	assert.Equal(t, fakeErr, err)
}
