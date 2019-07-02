package rs_test

import (
	"context"
	"fmt"

	rs "github.com/jjeffcaii/reactor-go"
	"github.com/jjeffcaii/reactor-go/flux"
	"github.com/jjeffcaii/reactor-go/scheduler"
)

func Example() {
	gen := func(ctx context.Context, sink flux.Sink) {
		for i := 0; i < 10; i++ {
			v := i
			sink.Next(v)
		}
		sink.Complete()
	}
	done := make(chan struct{})
	flux.Create(gen).
		Filter(func(i interface{}) bool {
			return i.(int)%2 == 0
		}).
		Map(func(i interface{}) interface{} {
			return fmt.Sprintf("#HELLO_%04d", i.(int))
		}).
		SubscribeOn(scheduler.Elastic()).
		Subscribe(context.Background(), rs.NewSubscriber(
			rs.OnNext(func(s rs.Subscription, v interface{}) {
				fmt.Println("next:", v)
			}),
			rs.OnComplete(func() {
				close(done)
			}),
		))
	<-done
}
