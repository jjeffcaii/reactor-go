package rs

import (
	"context"
	"log"
	"testing"
)

func TestNewMono(t *testing.T) {
	mono := NewMono(func(producer MonoProducer) {
		producer.Success("foobar")
	})
	mono.
		PublishOn(Elastic()).
		SubscribeOn(Elastic()).
		Subscribe(context.Background(), OnNext(func(ctx context.Context, sub Subscription, v interface{}) {
			log.Println("next1:", v)
		}), OnNext(func(ctx context.Context, sub Subscription, v interface{}) {
			log.Println("next2:", v)
		}))
}
