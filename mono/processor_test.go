package mono

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestProcessor(t *testing.T) {
	p := CreateProcessor()

	time.AfterFunc(100*time.Millisecond, func() {
		p.Success(333)
	})

	v, err := p.
		Map(func(i interface{}) interface{} {
			return i.(int) * 2
		}).
		Block(context.Background())
	assert.NoError(t, err, "block failed")
	assert.Equal(t, 666, v.(int), "bad result")

	var actual int
	p.
		DoOnNext(func(v interface{}) {
			actual = v.(int)
		}).
		Subscribe(context.Background())
	assert.Equal(t, 333, actual, "bad result")
}
