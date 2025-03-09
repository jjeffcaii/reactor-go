package reactor_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/jjeffcaii/reactor-go"
)

func TestSignalType_String(t *testing.T) {
	for _, v := range []struct {
		Signal reactor.SignalType
		String string
	}{
		{reactor.SignalTypeCancel, "CANCEL"},
		{reactor.SignalTypeComplete, "COMPLETE"},
		{reactor.SignalTypeDefault, "DEFAULT"},
		{reactor.SignalTypeError, "ERROR"},
	} {
		assert.Equal(t, v.String, v.Signal.String(), "should be same string")
	}
}
