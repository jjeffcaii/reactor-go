package mono_test

import (
	"context"
	"testing"
	"time"

	"github.com/jjeffcaii/reactor-go"
	"github.com/jjeffcaii/reactor-go/mono"
	"github.com/stretchr/testify/assert"
)

func TestFlatMap(t *testing.T) {
	const fakeItem = "fake item"
	actual, err := mono.Just(123).
		FlatMap(func(_ reactor.Any) mono.Mono {
			return mono.Delay(300 * time.Millisecond).
				Map(func(_ reactor.Any) (reactor.Any, error) {
					return fakeItem, nil
				})
		}).
		Block(context.Background())
	assert.NoError(t, err, "should not return error")
	assert.Equal(t, fakeItem, actual)
}
