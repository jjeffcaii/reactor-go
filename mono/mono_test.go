package mono_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/jjeffcaii/reactor-go"
	"github.com/jjeffcaii/reactor-go/mono"
)

func BenchmarkNative(b *testing.B) {
	for i := 0; i < b.N; i++ {
		time.Now().Add(-24 * time.Hour)
	}
}

func BenchmarkMono(b *testing.B) {
	for i := 0; i < b.N; i++ {
		mono.Just(time.Now()).
			Map(func(i interface{}) interface{} {
				return i.(time.Time).Add(-24 * time.Hour)
			}).
			Subscribe(context.Background())
	}
}

func Example() {
	gen := func(ctx context.Context, sink mono.Sink) {
		sink.Success("World")
	}
	mono.
		Create(gen).
		Map(func(i interface{}) interface{} {
			return "Hello " + i.(string) + "!"
		}).
		DoOnNext(func(s rs.Subscription, v interface{}) {
			fmt.Println(v)
		}).
		Subscribe(context.Background())
}

// Should print
// Hello World!
