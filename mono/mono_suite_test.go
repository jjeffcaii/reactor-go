package mono_test

import (
	"context"
	"errors"
	"testing"

	"github.com/jjeffcaii/reactor-go"
	"github.com/jjeffcaii/reactor-go/mono"
	"github.com/stretchr/testify/assert"
)

func TestName(t *testing.T) {
	t.Run("just", func(t *testing.T) {
		runPanic(mono.Just(1), t)
	})
}

func runMap(m mono.Mono, t *testing.T) {
	m.
		Map(func(i interface{}) interface{} {
			return i.(int) * 2
		}).
		DoOnNext(func(v interface{}) {
			assert.Equal(t, 2, v.(int), "bad doOnNext result")
		}).
		Subscribe(context.Background(), rs.OnNext(func(v interface{}) {
			assert.Equal(t, 2, v.(int), "bad onNext result")
		}))

}

func runPanic(m mono.Mono, t *testing.T) {
	m.
		Map(func(i interface{}) interface{} {
			panic(errors.New("foobar"))
		}).
		DoOnError(func(e error) {
			assert.Error(t, e, "should catch error")
		}).
		Subscribe(context.Background())
}
