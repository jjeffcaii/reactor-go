package flux_test

import (
  "context"
  "fmt"
  "log"
  "testing"
  "time"

  rs "github.com/jjeffcaii/reactor-go"
  "github.com/jjeffcaii/reactor-go/flux"
  "github.com/jjeffcaii/reactor-go/hooks"
  "github.com/stretchr/testify/assert"
)

func TestUnicastProcessor(t *testing.T) {
  hooks.OnNextDrop(func(v interface{}) {
    fmt.Println("dropped:", v)
  })
  p := flux.NewUnicastProcessor()

  time.AfterFunc(10*time.Millisecond, func() {
    p.Next(1)
  })

  time.AfterFunc(15*time.Millisecond, func() {
    p.Next(2)
  })
  time.AfterFunc(20*time.Millisecond, func() {
    p.Next(3)
    p.Complete()
  })

  results := make(chan int)

  var su rs.Subscription
  p.
    DoOnNext(func(v interface{}) {
      time.Sleep(30 * time.Millisecond)
      log.Println("onNext:", v)
      results <- v.(int)
      su.Request(1)
    }).
    DoOnRequest(func(n int) {
      log.Println("request:", n)
    }).
    DoOnComplete(func() {
      log.Println("complete")
      close(results)
    }).
    Subscribe(context.Background(), rs.OnSubscribe(func(s rs.Subscription) {
      su = s
      su.Request(1)
    }))

  done := make(chan struct{})
  go func() {
    defer close(done)
    var sum int
    for v := range results {
      sum += v
    }
    assert.Equal(t, 6, sum, "bad sum result")
  }()
  <-done
}
