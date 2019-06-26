package mono

import (
	"context"
	"fmt"
	"log"
	"testing"

	"github.com/jjeffcaii/reactor-go"
)

func TestNew(t *testing.T) {
	mono := New(func(producer Sink) {
		producer.Success(1)
	})
	mono.
		Map(func(i interface{}) interface{} {
			return i.(int) * 2
		}).
		Map(func(i interface{}) interface{} {
			return fmt.Sprintf("mapped_%04d", i)
		}).
		Subscribe(context.Background(), rs.OnNext(func(s rs.Subscription, v interface{}) {
			log.Println("next1:", v)
		}), rs.OnNext(func(sub rs.Subscription, v interface{}) {
			log.Println("next2:", v)
		}))
}
