package scheduler_test

import (
	"fmt"
	"testing"

	"github.com/jjeffcaii/reactor-go/scheduler"
	"github.com/stretchr/testify/assert"
)

func TestImmediate(t *testing.T) {
	fmt.Println(scheduler.Immediate().Name())
	defer scheduler.Immediate().Close()
	assert.NotPanics(t, func() {
		scheduler.Immediate().Worker().Do(func() {
			// noop
		})
	})
}
