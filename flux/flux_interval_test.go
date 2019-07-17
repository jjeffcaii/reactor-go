package flux_test

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	rs "github.com/jjeffcaii/reactor-go"
	"github.com/jjeffcaii/reactor-go/flux"
	"github.com/stretchr/testify/assert"
)

func TestInterval(t *testing.T) {
	done := make(chan struct{})
	var amount int32
	flux.Interval(100*time.Millisecond).
		DoFinally(func(s rs.SignalType) {
			close(done)
		}).
		Subscribe(context.Background(),
			rs.OnNext(func(v interface{}) {
				atomic.AddInt32(&amount, 1)
			}),
			rs.OnSubscribe(func(su rs.Subscription) {
				su.Request(3)
			}),
		)
	<-done
	assert.Equal(t, 3, int(amount), "bad amount")
}
