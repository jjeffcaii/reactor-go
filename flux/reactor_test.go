package flux_test

import (
	"context"
	"fmt"

	"github.com/jjeffcaii/reactor-go"
	. "github.com/jjeffcaii/reactor-go/flux"
	"github.com/jjeffcaii/reactor-go/scheduler"
)

func Example() {
	gen := func(ctx context.Context, sink Sink) {
		for i := 0; i < 10; i++ {
			v := i
			sink.Next(v)
		}
		sink.Complete()
	}
	done := make(chan struct{})
	Create(gen).
		Filter(func(i interface{}) bool {
			return i.(int)%2 == 0
		}).
		Map(func(i interface{}) interface{} {
			return fmt.Sprintf("#HELLO_%04d", i.(int))
		}).
		SubscribeOn(scheduler.Elastic()).
		Subscribe(context.Background(),
			rs.OnSubscribe(func(s rs.Subscription) {
				s.Request(1)
			}),
			rs.OnNext(func(s rs.Subscription, v interface{}) {
				fmt.Println("next:", v)
				s.Request(1)
			}),
			rs.OnComplete(func() {
				close(done)
			}),
		)
	<-done
}
