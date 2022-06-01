package internal_test

import (
	"testing"

	"github.com/jjeffcaii/reactor-go/internal"
	"github.com/stretchr/testify/assert"
)

func TestEmptySubscription(t *testing.T) {
	assert.NotPanics(t, func() {
		internal.EmptySubscription.Cancel()
		internal.EmptySubscription.Request(1)
	})
}

func TestSafeCloseDone(t *testing.T) {
	c := make(chan struct{})
	assert.True(t, internal.SafeCloseDone(c))
	assert.False(t, internal.SafeCloseDone(c))
}
