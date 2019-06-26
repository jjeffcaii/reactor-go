package rs

import (
	"context"
	"fmt"
	"log"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestFlux_Simple(t *testing.T) {

	flux := NewFlux(func(ctx context.Context, producer FluxSink) {
		for i := 0; i < 10; i++ {
			time.Sleep(100 * time.Millisecond)
			producer.Next(fmt.Sprintf("MSG_%04d", i))
		}
		producer.Complete()
	})
	flux.
		Subscribe(context.Background(), OnNext(func(sub Subscription, v interface{}) {
			log.Println("onNext:", v)
		}), OnError(func(err error) {
			log.Println("onError:", err)
			require.NoError(t, err)
		}))
}

func TestFlux_Map(t *testing.T) {
	f := NewFlux(func(ctx context.Context, producer FluxSink) {
		for i := 0; i < 100; i++ {
			_ = producer.Next(i)
		}
		producer.Complete()
	})

	f.
		Map(func(i interface{}) interface{} {
			return i.(int) * 2
		}).
		Filter(func(i interface{}) bool {
			return i.(int)%4 != 0
		}).
		Map(func(i interface{}) interface{} {
			return fmt.Sprintf("message_%04d", i.(int))
		}).
		Subscribe(context.Background(), OnNext(func(s Subscription, v interface{}) {
			log.Println("next:", v)
		}))

}

func TestFlux_Request(t *testing.T) {
	f := NewFlux(func(ctx context.Context, producer FluxSink) {
		for i := 0; i < 100; i++ {
			_ = producer.Next(fmt.Sprintf("message_%04d", i))
		}
		producer.Complete()
	})

	f.Subscribe(
		context.Background(),
		OnRequest(func(n int) {
			log.Println("request:", n)
		}),
		OnSubscribe(func(s Subscription) {
			s.Request(1)
		}),
		OnNext(func(s Subscription, v interface{}) {
			log.Println("next:", v)
			s.Request(1)
		}),
		OnComplete(func() {
			log.Println("finish")
		}),
	)

}
