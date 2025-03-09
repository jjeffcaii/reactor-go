package main

import (
	"context"
	"log"

	"github.com/jjeffcaii/reactor-go"
	mono "github.com/jjeffcaii/reactor-go/mono2"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	mono.Create[int](func(ctx context.Context, sink mono.Sink[int]) {
		sink.Success(1234)
	}).Subscribe(ctx,
		reactor.OnNext[int](func(ctx context.Context, item int) error {
			log.Println("onNext:", item)
			return nil
		}),
		reactor.OnError[int](func(ctx context.Context, err error) {
			log.Println("onError:", err)
		}),
	)
}
