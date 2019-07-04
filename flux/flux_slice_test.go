package flux

import (
	"context"
	"fmt"
	"log"
	"testing"

	rs "github.com/jjeffcaii/reactor-go"
	"github.com/jjeffcaii/reactor-go/scheduler"
)

func TestJust2(t *testing.T) {
	done := make(chan struct{})
	Just(1, 2, 3, 4).
		Map(func(i interface{}) interface{} {
			return i.(int) * 2
		}).
		Filter(func(i interface{}) bool {
			return i.(int) > 4
		}).
		Map(func(i interface{}) interface{} {
			return fmt.Sprintf("A%04d", i.(int))
		}).
		SubscribeOn(scheduler.Elastic()).
		Subscribe(context.Background(),
			rs.OnNext(func(s rs.Subscription, i interface{}) {
				log.Println("next:", i)
			}),
			rs.OnComplete(func() {
				close(done)
			}),
		)
	<-done
}
