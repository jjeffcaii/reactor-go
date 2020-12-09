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
