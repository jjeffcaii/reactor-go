package flux_test

import (
	"context"
	"log"
	"sync/atomic"
	"testing"
	"time"

	rs "github.com/jjeffcaii/reactor-go"
	"github.com/jjeffcaii/reactor-go/flux"
	"github.com/stretchr/testify/assert"
)

func TestInterval(t *testing.T) {
	var amount int32
	flux.Interval(100*time.Millisecond).Subscribe(context.Background(),
		rs.OnNext(func(v interface{}) {
			log.Println("next:", v)
			atomic.AddInt32(&amount, 1)
		}),
		rs.OnSubscribe(func(su rs.Subscription) {
			su.Request(3)
		}),
	)
	time.Sleep(500 * time.Millisecond)
	assert.Equal(t, 3, int(amount), "bad amount")
}
