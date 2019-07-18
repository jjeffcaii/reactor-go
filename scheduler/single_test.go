package scheduler

import (
	"log"
	"sync/atomic"
	"testing"
	"time"
)

func TestSingle(t *testing.T) {
	done := make(chan struct{})
	var totals int32
	for range [100]struct{}{} {
		Single().Worker().Do(func() {
			log.Println(time.Now())
			if atomic.AddInt32(&totals, 1) == 100 {
				close(done)
			}
		})
	}
	<-done
}
