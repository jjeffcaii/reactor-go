package rs

import (
	"context"
	"fmt"
	"log"
	"testing"
)

func TestName(t *testing.T) {
	NewFlux(func(producer Producer) {
		for i := 0; i < 100; i++ {
			producer.Next(fmt.Sprintf("MSG_%04d", i))
		}
		producer.Complete()
	}).Subscribe(context.Background(), OnSubscribe(func(subscription Subscription) {
		subscription.Cancel()
	}), OnNext(func(v interface{}) {
		log.Println("next:", v)
	}))

}
