package flux_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/jjeffcaii/reactor-go"
	"github.com/jjeffcaii/reactor-go/flux"
	"github.com/stretchr/testify/assert"
)

func TestSwitchOnFirst(t *testing.T) {
	// greater than first element and multiply 2
	origin := []Any{3, 8, 7, 6, 5, 4, 3, 2, 1}

	var expected []int
	for _, it := range origin {
		if it.(int) > origin[0].(int) {
			expected = append(expected, it.(int)*2)
		}
	}
	var actual []int
	flux.Just(origin...).
		SwitchOnFirst(func(s flux.Signal, f flux.Flux) flux.Flux {
			if first, ok := s.Value(); ok {
				return f.
					Filter(func(i Any) bool {
						return i.(int) > first.(int)
					}).
					Map(func(i Any) (Any, error) {
						return i.(int) * 2, nil
					})
			}
			return f
		}).
		DoOnNext(func(v Any) error {
			actual = append(actual, v.(int))
			fmt.Println("next:", v)
			return nil
		}).
		DoFinally(func(s reactor.SignalType) {
			fmt.Println("finally:", s)
		}).
		Subscribe(context.Background())
	assert.Equal(t, expected, actual, "bad result")
}
