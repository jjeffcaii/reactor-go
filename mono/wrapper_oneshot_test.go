package mono_test

import (
	"context"
	"testing"

	"github.com/jjeffcaii/reactor-go"
	"github.com/jjeffcaii/reactor-go/mono"
	"github.com/stretchr/testify/assert"
)

func TestOneshotWrapper(t *testing.T) {
	_, err := mono.JustOneshot(1).
		Map(func(any reactor.Any) (reactor.Any, error) {
			return any.(int) * 2, nil
		}).
		FlatMap(func(any reactor.Any) mono.Mono {
			return mono.JustOneshot(any)
		}).
		DoOnNext(func(v reactor.Any) error {
			assert.Equal(t, 2, v)
			return nil
		}).
		Filter(func(any reactor.Any) bool {
			return any.(int) > 0
		}).
		DoOnCancel(func() {
			assert.Fail(t, "unreachable!")
		}).
		Map(func(any reactor.Any) (reactor.Any, error) {
			return nil, fakeErr
		}).
		Block(context.Background())
	assert.Equal(t, fakeErr, err)
}
