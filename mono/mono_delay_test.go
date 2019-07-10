package mono_test

import (
	"context"
	"testing"
	"time"

	"github.com/jjeffcaii/reactor-go/mono"
	"github.com/stretchr/testify/assert"
)

func TestDelay(t *testing.T) {
	begin := time.Now()
	v, err := mono.Delay(1 * time.Second).Block(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, 1, int(time.Since(begin).Seconds()))
	assert.Equal(t, int64(0), v)
}
