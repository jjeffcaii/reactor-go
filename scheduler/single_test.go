package scheduler_test

import (
	"fmt"
	"sync"
	"testing"

	"github.com/jjeffcaii/reactor-go/scheduler"
	"github.com/stretchr/testify/assert"
)

const totals = 100

func TestSingle(t *testing.T) {
	single := scheduler.Single()
	fmt.Println(single.Name())

	wg := sync.WaitGroup{}
	wg.Add(totals)
	for range [totals]struct{}{} {
		assert.NoError(t, single.Worker().Do(func() {
			wg.Done()
		}), "should not return error")
	}
	wg.Wait()

	assert.NoError(t, single.Close(), "close should not return error")
	noop := func() {
		// noop
	}
	assert.Error(t, single.Worker().Do(noop), "should return error after closing scheduler")
}
