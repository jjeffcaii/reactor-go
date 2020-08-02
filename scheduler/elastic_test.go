package scheduler_test

import (
	"fmt"
	"sync"
	"testing"

	"github.com/jjeffcaii/reactor-go/scheduler"
	"github.com/stretchr/testify/assert"
)

func TestElastic(t *testing.T) {
	fmt.Println(scheduler.Elastic().Name())
	done := make(chan int)
	scheduler.Elastic().Worker().Do(func() {
		done <- 1
	})
	assert.Equal(t, 1, <-done)
}

func TestNewElastic(t *testing.T) {
	const size = 100
	sc := scheduler.NewElastic(size)

	assert.NotPanics(t, func() {
		wg := sync.WaitGroup{}
		wg.Add(size * 2)
		for i := 0; i < size*2; i++ {
			sc.Worker().Do(func() {
				wg.Done()
			})
		}
		wg.Wait()
	})
	assert.NoError(t, sc.Close())

	assert.Panics(t, func() {
		sc.Worker().Do(func() {
			// noop
		})
	}, "should panic after closing scheduler")

}
