package mono_test

import (
	"context"
	"log"
	"testing"

	rs "github.com/jjeffcaii/reactor-go"
	"github.com/jjeffcaii/reactor-go/mono"
	"github.com/stretchr/testify/assert"
)

func TestSwitchIfEmpty(t *testing.T) {
	m := mono.Create(func(ctx context.Context, sink mono.Sink) {
		sink.Success(nil)
	})
	m.SwitchIfEmpty(mono.Just(333)).
		Map(func(i interface{}) interface{} {
			return i.(int) * 2
		}).
		DoOnNext(func(s rs.Subscription, v interface{}) {
			assert.Equal(t, 666, v.(int), "bad result")
			log.Println("next:", 666)
		}).
		Subscribe(context.Background())

}
