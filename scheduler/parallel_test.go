package scheduler_test

import (
	"testing"

	"github.com/jjeffcaii/reactor-go/scheduler"
	"github.com/stretchr/testify/assert"
)

func TestParallel(t *testing.T) {
	defer scheduler.Parallel().Close()
	done := make(chan int)
	scheduler.Parallel().Worker().Do(func() {
		done <- 1
	})
	assert.Equal(t, 1, <-done)
}
