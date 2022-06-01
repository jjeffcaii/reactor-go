package mono

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestProcessorPool_PutWithNil(t *testing.T) {
	assert.NotPanics(t, func() {
		globalProcessorPool.put(nil)
	})
	assert.NotPanics(t, func() {
		globalProcessorSubscriberPool.put(nil)
	})
}
