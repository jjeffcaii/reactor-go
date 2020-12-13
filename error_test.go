package reactor_test

import (
	"testing"

	"github.com/jjeffcaii/reactor-go"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

func TestContextError(t *testing.T) {
	fakeErr := errors.New("fake error")
	err := reactor.NewContextError(fakeErr)
	assert.True(t, errors.Cause(err) == fakeErr, "should cause by fakeErr")
	assert.Equal(t, fakeErr.Error(), err.Error(), "bad error string")
}

func TestIsCancelledError(t *testing.T) {
	fakeErr := errors.New("fake error")
	assert.False(t, reactor.IsCancelledError(nil))
	assert.False(t, reactor.IsCancelledError(fakeErr))
	assert.True(t, reactor.IsCancelledError(reactor.ErrSubscribeCancelled))
	ctxErr := reactor.NewContextError(errors.New("fake context error"))
	assert.True(t, reactor.IsCancelledError(ctxErr))
}
