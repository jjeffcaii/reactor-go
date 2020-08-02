package scheduler_test

import (
	"testing"

	"github.com/jjeffcaii/reactor-go/scheduler"
	"github.com/stretchr/testify/assert"
)

func TestImmediate(t *testing.T) {
	defer scheduler.Immediate().Close()
	assert.NotPanics(t, func() {
		scheduler.Immediate().Worker().Do(func() {
			// noop
		})
	})
}
