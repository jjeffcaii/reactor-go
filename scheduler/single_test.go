package scheduler_test

import (
	"sync"
	"testing"

	"github.com/jjeffcaii/reactor-go/scheduler"
	"github.com/stretchr/testify/assert"
)

const totals = 100

func TestSingle(t *testing.T) {
	single := scheduler.Single()

	assert.NotPanics(t, func() {
		wg := sync.WaitGroup{}
		wg.Add(totals)
		for range [totals]struct{}{} {
			single.Worker().Do(func() {
				wg.Done()
			})
		}
		wg.Wait()
	})
	assert.NoError(t, single.Close(), "close should not return error")

	assert.Panics(t, func() {
		single.Worker().Do(func() {
			// noop
		})
	}, "should panic after closing scheduler")
}
