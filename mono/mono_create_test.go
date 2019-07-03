package mono_test

import (
	"context"
	"log"
	"testing"

	rs "github.com/jjeffcaii/reactor-go"
	"github.com/jjeffcaii/reactor-go/mono"
	"github.com/jjeffcaii/reactor-go/scheduler"
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

func TestMonoCreate_Filter(t *testing.T) {
	var next, complete bool
	mono.
		Create(func(i context.Context, sink mono.Sink) {
			sink.Success(9)
		}).
		Map(func(i interface{}) interface{} {
			return i.(int) * 2
		}).
		Filter(func(i interface{}) bool {
			return i.(int) < 10
		}).
		Subscribe(context.Background(), rs.NewSubscriber(
			rs.OnNext(func(s rs.Subscription, v interface{}) {
				log.Println("next:", v)
				next = true
			}),
			rs.OnComplete(func() {
				log.Println("complete")
				complete = true
			}),
		))
	assert.False(t, next, "bad next")
	assert.True(t, complete, "bad complete")
}

func TestMonoCreate_SubscribeOn(t *testing.T) {
	gen := func(i context.Context, sink mono.Sink) {
		sink.Success(1)
	}
	done := make(chan struct{})
	mono.Create(gen).
		SubscribeOn(scheduler.Elastic()).
		Subscribe(context.Background(), rs.NewSubscriber(
			rs.OnNext(func(s rs.Subscription, v interface{}) {
				log.Println("next:", v)
			}),
			rs.OnComplete(func() {
				close(done)
			}),
		))
	<-done
}
