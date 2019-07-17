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
	origin := []interface{}{3, 8, 7, 6, 5, 4, 3, 2, 1}

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
					Filter(func(i interface{}) bool {
						return i.(int) > first.(int)
					}).
					Map(func(i interface{}) interface{} {
						return i.(int) * 2
					})
			}
			return f
		}).
		DoOnNext(func(v interface{}) {
			actual = append(actual, v.(int))
			fmt.Println("next:", v)
		}).
		DoFinally(func(s rs.SignalType) {
			fmt.Println("finally:", s)
		}).
		Subscribe(context.Background())
	assert.Equal(t, expected, actual, "bad result")
}
