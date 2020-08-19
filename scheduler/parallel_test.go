package scheduler_test

import (
	"fmt"
	"testing"

	"github.com/jjeffcaii/reactor-go/scheduler"
	"github.com/stretchr/testify/assert"
)

func TestParallel(t *testing.T) {
	fmt.Println(scheduler.Parallel().Name())
	defer scheduler.Parallel().Close()
	done := make(chan int)
	_ = scheduler.Parallel().Worker().Do(func() {
		done <- 1
	})
	assert.Equal(t, 1, <-done)
}
