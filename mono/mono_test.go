package mono

import (
	"context"
	"fmt"
	"log"
	"testing"

	"github.com/jjeffcaii/reactor-go"
	"github.com/stretchr/testify/assert"
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

func TestJust(t *testing.T) {
	Just(1).
		MapInt(func(i int) interface{} {
			return fmt.Sprintf("%04d", i*2)
		}).
		MapString(func(s string) interface{} {
			return "A" + s
		}).
		Subscribe(context.Background(), rs.OnNextString(func(s rs.Subscription, i string) {
			assert.Equal(t, "A0002", i, "bad result")
		}))

}
