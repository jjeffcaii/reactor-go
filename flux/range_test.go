package flux_test

import (
	"context"
	"testing"

	rs "github.com/jjeffcaii/reactor-go"
	"github.com/jjeffcaii/reactor-go/flux"
	"github.com/stretchr/testify/assert"
)

func TestRangeFast(t *testing.T) {
	const totals = 100
	var sum int64
	flux.
		Range(0, totals).
		DoOnNext(func(v interface{}) {
			sum += v.(int64)
		}).
		Subscribe(context.Background())

	var should int64
	for i := 0; i < totals; i++ {
		should += int64(i)
	}
	assert.Equal(t, should, sum, "bad sum")
}

func TestRangeSlow(t *testing.T) {
	const totals = 10
	var sum int64
	var requests int
	var su rs.Subscription
	flux.
		Range(0, totals).
		DoOnRequest(func(n int) {
			requests++
		}).
		DoOnNext(func(v interface{}) {
			sum += v.(int64)
			su.Request(1)
		}).
		Subscribe(context.Background(), rs.OnSubscribe(func(s rs.Subscription) {
			su = s
			su.Request(1)
		}))
	var should int64
	for i := 0; i < totals; i++ {
		should += int64(i)
	}
	assert.Equal(t, should, sum, "bad sum")
	assert.Equal(t, totals+1, requests, "bad requests")
}
