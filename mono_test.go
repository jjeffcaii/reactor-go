package rs

import (
	"context"
	"fmt"
	"log"
	"testing"
)

func TestNewMono(t *testing.T) {
	mono := NewMono(func(producer MonoProducer) {
		producer.Success(1)
	})
	mono.
		Map(func(i interface{}) interface{} {
			return i.(int) * 2
		}).
		Map(func(i interface{}) interface{} {
			return fmt.Sprintf("mapped_%04d", i)
		}).
		Subscribe(context.Background(), OnNext(func(ctx context.Context, sub Subscription, v interface{}) {
			log.Println("next1:", v)
		}), OnNext(func(ctx context.Context, sub Subscription, v interface{}) {
			log.Println("next2:", v)
		}))
}
