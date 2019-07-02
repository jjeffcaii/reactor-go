package mono_test

import (
	"context"
	"fmt"

	rs "github.com/jjeffcaii/reactor-go"
	"github.com/jjeffcaii/reactor-go/mono"
)

func Example() {
	gen := func(ctx context.Context, sink mono.Sink) {
		sink.Success("World")
	}
	mono.
		Create(gen).
		Map(func(i interface{}) interface{} {
			return "Hello " + i.(string) + "!"
		}).
		Subscribe(context.Background(), rs.NewSubscriber(
			rs.OnNext(func(s rs.Subscription, v interface{}) {
				fmt.Println(v)
			}),
		))
}

// Should print
// Hello World!
