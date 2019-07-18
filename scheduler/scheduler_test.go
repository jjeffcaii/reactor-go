package scheduler_test

import (
	"log"
	"sync/atomic"
	"testing"
	"time"

	"github.com/jjeffcaii/reactor-go/scheduler"
)

const totals = 100

func TestSingle(t *testing.T) {
	done := make(chan struct{})
	lefts := int64(totals)
	for range [totals]struct{}{} {
		scheduler.Single().Worker().Do(func() {
			log.Println(time.Now())
			if atomic.AddInt64(&lefts, -1) == 0 {
				close(done)
			}
		})
	}
	<-done
}
