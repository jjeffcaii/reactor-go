package mono_test

import (
	"context"
	"testing"
	"time"

	"github.com/jjeffcaii/reactor-go/mono"
	"github.com/stretchr/testify/assert"
)

func TestDelayElement(t *testing.T) {
	delay := 100 * time.Millisecond
	start := time.Now()
	v, err := mono.Just(1).DelayElement(delay).Block(context.Background())
	assert.NoError(t, err, "should not return error")
	assert.Equal(t, 1, v, "value doesn't match")
	assert.True(t, int64(time.Since(start)) > int64(delay), "should greater than delay duration")

	start = time.Now()
	v, err = mono.Empty().DelayElement(delay).Block(context.Background())
	assert.NoError(t, err, "should not return error")
	assert.Nil(t, v, "should be nil if mono is empty")
	assert.True(t, int64(time.Since(start)) < int64(delay), "should less than delay duration")

	start = time.Now()
	_, err = mono.Error(fakeErr).DelayElement(delay).Block(context.Background())
	assert.Equal(t, fakeErr, err, "should be fake error")
	assert.True(t, int64(time.Since(start)) < int64(delay), "should less than delay duration")
}
