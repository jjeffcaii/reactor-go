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
	assert.True(t, errors.Cause(err) == fakeErr)
}
