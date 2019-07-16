package flux

import (
	"context"
	"log"
	"testing"
	"time"

	rs "github.com/jjeffcaii/reactor-go"
)

func TestUnicastProcessor(t *testing.T) {
	p := NewUnicastProcessor()

	time.AfterFunc(100*time.Millisecond, func() {
		p.Next(1)
	})

	time.AfterFunc(150*time.Millisecond, func() {
		p.Next(2)
	})
	time.AfterFunc(200*time.Millisecond, func() {
		p.Next(2)
		p.Complete()
	})

	done := make(chan struct{})
	var su rs.Subscription
	p.DoOnNext(func(v interface{}) {
		log.Println("onNext:", v)
		su.Request(1)
	}).DoOnRequest(func(n int) {
		log.Println("request:", n)
	}).DoOnComplete(func() {
		log.Println("complete")
		close(done)
	}).Subscribe(context.Background(), rs.OnSubscribe(func(s rs.Subscription) {
		su = s
		su.Request(1)
	}))
	<-done
}
