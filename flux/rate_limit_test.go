package flux_test

import (
	"context"
	"log"
	"testing"
	"time"

	"github.com/jjeffcaii/reactor-go/flux"
)

func TestRateLimit(t *testing.T) {
	_, _ = flux.Interval(10*time.Millisecond).
		DoOnNext(func(v interface{}) {
			log.Println("ON_NEXT:", v)
		}).
		DoOnRequest(func(n int) {
			log.Println("ON_REQUEST:", n)
		}).
		Take(10).
		RateLimit(5, 2).
		BlockLast(context.Background())
}
