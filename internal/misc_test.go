package internal_test

import (
	"errors"
	"testing"

	"github.com/jjeffcaii/reactor-go/internal"
	"github.com/stretchr/testify/assert"
)

func TestTryRecoverError(t *testing.T) {
	fakeErr := errors.New("fake error")
	assert.Equal(t, fakeErr, internal.TryRecoverError(fakeErr))
	assert.Error(t, internal.TryRecoverError("fake error"))
	assert.Error(t, internal.TryRecoverError(123))
}

func TestEmptySubscription(t *testing.T) {
	assert.NotPanics(t, func() {
		internal.EmptySubscription.Cancel()
		internal.EmptySubscription.Request(1)
	})
}
