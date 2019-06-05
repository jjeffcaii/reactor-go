package rs

import (
	"context"
	"fmt"
	"log"
	"testing"
	"time"
)

func TestFlux_Simple(t *testing.T) {
	flux := NewFlux(func(ctx context.Context, producer Producer) {
		for i := 0; i < 10; i++ {
			time.Sleep(100 * time.Millisecond)
			producer.Next(fmt.Sprintf("MSG_%04d", i))
		}
		producer.Complete()
	})
	flux.
		Subscribe(context.Background(), OnNext(func(ctx context.Context, sub Subscription, v interface{}) {
			log.Println("onNext:", v)
		}), OnError(func(ctx context.Context, err error) {
			log.Println("onError:", err)
		}))
}

func TestFlux_Map(t *testing.T) {
	f := NewFlux(func(ctx context.Context, producer Producer) {
		for i := 0; i < 100; i++ {
			_ = producer.Next(i)
		}
		producer.Complete()
	})

	f.
		Map(func(i interface{}) interface{} {
			return i.(int) * 2
		}).
		Map(func(i interface{}) interface{} {
			return fmt.Sprintf("message_%04d", i.(int))
		}).
		Subscribe(context.Background(), OnNext(func(ctx context.Context, s Subscription, v interface{}) {
			log.Println("next:", v)
		}))

}

func TestFlux_Request(t *testing.T) {
	f := NewFlux(func(ctx context.Context, producer Producer) {
		for i := 0; i < 100; i++ {
			_ = producer.Next(fmt.Sprintf("message_%04d", i))
		}
		producer.Complete()
	})

	f.Subscribe(
		context.Background(),
		OnRequest(func(ctx context.Context, n int) {
			log.Println("request:", n)
		}),
		OnSubscribe(func(ctx context.Context, s Subscription) {
			s.Request(1)
		}),
		OnNext(func(ctx context.Context, s Subscription, v interface{}) {
			log.Println("next:", v)
			s.Request(1)
		}),
		OnComplete(func(ctx context.Context) {
			log.Println("finish")
		}),
	)

}
