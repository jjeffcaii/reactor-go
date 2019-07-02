package mono_test

import (
	"context"
	"log"
	"testing"

	rs "github.com/jjeffcaii/reactor-go"
	"github.com/jjeffcaii/reactor-go/mono"
	"github.com/stretchr/testify/assert"
)

func TestMonoCreate_Subscribe(t *testing.T) {
	var complete bool
	mono.
		Create(func(ctx context.Context, sink mono.Sink) {
			sink.Success(1)
		}).
		Map(func(i interface{}) interface{} {
			return i.(int) * 2
		}).
		Map(func(i interface{}) interface{} {
			return i.(int) * 2
		}).
		Subscribe(context.Background(), rs.NewSubscriber(
			rs.OnNext(func(s rs.Subscription, v interface{}) {
				log.Println("next:", v)
				assert.Equal(t, 4, v, "bad result")
			}),
			rs.OnComplete(func() {
				log.Println("complete")
				complete = true
			}),
		))
	assert.True(t, complete, "not complete")
}
