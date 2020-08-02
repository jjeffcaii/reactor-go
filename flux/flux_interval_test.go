package flux_test

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/jjeffcaii/reactor-go"
	"github.com/jjeffcaii/reactor-go/flux"
	"github.com/stretchr/testify/assert"
)

type Any = reactor.Any

func TestInterval(t *testing.T) {
	done := make(chan struct{})
	var amount int32
	flux.Interval(100*time.Millisecond).
		DoFinally(func(s reactor.SignalType) {
			close(done)
		}).
		Subscribe(context.Background(),
			reactor.OnNext(func(v Any) error {
				atomic.AddInt32(&amount, 1)
				return nil
			}),
			reactor.OnSubscribe(func(su reactor.Subscription) {
				su.Request(3)
			}),
		)
	<-done
	assert.Equal(t, 3, int(amount), "bad amount")
}
