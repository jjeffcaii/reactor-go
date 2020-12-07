package mono_test

import (
	"context"
	"testing"

	"github.com/jjeffcaii/reactor-go"
	"github.com/jjeffcaii/reactor-go/mono"
	"github.com/stretchr/testify/assert"
)

func TestDoFinally(t *testing.T) {
	var seq []int
	mono.Just(77778888).
		DoOnNext(func(v reactor.Any) error {
			seq = append(seq, 1)
			return nil
		}).
		DoFinally(func(s reactor.SignalType) {
			seq = append(seq, 3)
		}).
		DoOnNext(func(v reactor.Any) error {
			seq = append(seq, 2)
			return nil
		}).
		Subscribe(context.Background())
	assert.Equal(t, []int{1, 2, 3}, seq, "wrong execute order")
}
