package mono_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/jjeffcaii/reactor-go"
	"github.com/jjeffcaii/reactor-go/mono"
	"github.com/stretchr/testify/assert"
)

func TestDefaultIfError(t *testing.T) {
	v, err := mono.JustOneshot(15555).
		Map(func(any reactor.Any) (reactor.Any, error) {
			return any, nil
		}).
		Map(func(any reactor.Any) (reactor.Any, error) {
			return any, fmt.Errorf("trigger err")
		}).
		DoOnError(func(e error) {
		}).
		DefaultIfError(123).
		Map(func(any reactor.Any) (reactor.Any, error) {
			return any, nil
		}).
		Block(context.Background())
	assert.NoError(t, err, "err occurred")
	assert.Equal(t, 123, v.(int), "bad result")

	fakeErr := errors.New("fake error")
	v, err = mono.Error(fakeErr).DefaultIfError(1).Block(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, 1, v)

	v, err = mono.Just(1).DefaultIfError(2).Block(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, 1, v)
}
