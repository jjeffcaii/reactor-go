package flux_test

import (
	"context"
	"fmt"
	"log"
	"testing"

	"github.com/jjeffcaii/reactor-go"
	"github.com/jjeffcaii/reactor-go/flux"
	"github.com/jjeffcaii/reactor-go/scheduler"
	"github.com/stretchr/testify/assert"
)

func Example() {
	gen := func(ctx context.Context, sink flux.Sink) {
		for i := 0; i < 10; i++ {
			v := i
			sink.Next(v)
		}
		sink.Complete()
	}
	done := make(chan struct{})

	var su rs.Subscription
	flux.Create(gen).
		Filter(func(i interface{}) bool {
			return i.(int)%2 == 0
		}).
		Map(func(i interface{}) interface{} {
			return fmt.Sprintf("#HELLO_%04d", i.(int))
		}).
		SubscribeOn(scheduler.Elastic()).
		Subscribe(context.Background(),
			rs.OnSubscribe(func(s rs.Subscription) {
				su = s
				s.Request(1)
			}),
			rs.OnNext(func(v interface{}) {
				fmt.Println("next:", v)
				su.Request(1)
			}),
			rs.OnComplete(func() {
				close(done)
			}),
		)
	<-done
}

var testData = []int{1, 2, 3, 4, 5}

func TestSuite(t *testing.T) {
	var inputs []interface{}
	for _, value := range testData {
		inputs = append(inputs, value)
	}
	all := make(map[string]flux.Flux)
	all["just"] = flux.Just(inputs...)

	for k, v := range all {
		t.Run(fmt.Sprintf("%s_Request", k), func(t *testing.T) {
			testRequest(v, t)
		})
		t.Run(fmt.Sprintf("%s_Peek", k), func(t *testing.T) {
			testPeek(v, t)
		})
		t.Run(fmt.Sprintf("%s_Discard", k), func(t *testing.T) {
			testDiscard(v, t)
		})
	}
}

func testDiscard(f flux.Flux, t *testing.T) {
	var next, next2, discard []int
	f.
		DoOnNext(func(v interface{}) {
			next = append(next, v.(int))
		}).
		Filter(func(i interface{}) bool {
			return i.(int) > 3
		}).
		DoOnNext(func(v interface{}) {
			next2 = append(next2, v.(int))
		}).
		DoOnDiscard(func(i interface{}) {
			discard = append(discard, i.(int))
		}).
		Subscribe(context.Background())
	assert.Equal(t, testData, next, "bad next")
	assert.Equal(t, len(next), len(next2)+len(discard), "bad amount")
	assert.Equal(t, []int{4, 5}, next2, "bad next2")
	assert.Equal(t, []int{1, 2, 3}, discard, "bad discard")
}

func testPeek(f flux.Flux, t *testing.T) {
	var complete int
	var a, b []int
	var requests int
	var ss rs.Subscription
	f.
		DoOnNext(func(v interface{}) {
			log.Println("peek next:", v)
			a = append(a, v.(int))
		}).
		DoOnRequest(func(n int) {
			requests++
		}).
		DoOnComplete(func() {
			log.Println("peek complete")
			complete++
		}).
		Subscribe(context.Background(), rs.OnSubscribe(func(su rs.Subscription) {
			ss = su
			ss.Request(1)
		}), rs.OnNext(func(v interface{}) {
			b = append(b, v.(int))
			ss.Request(1)
		}))
	assert.Equal(t, b, a, "values doesn't match")
	assert.Equal(t, len(a)+1, requests, "bad requests")
	assert.Equal(t, 1, complete, "bad complete")
}

func testRequest(f flux.Flux, t *testing.T) {
	var ss rs.Subscription
	f.Subscribe(context.Background(), rs.OnSubscribe(func(su rs.Subscription) {
		ss = su
		su.Request(2)
	}), rs.OnNext(func(v interface{}) {
		log.Println("next:", v)
		ss.Request(2)
	}))

}

func TestEmpty(t *testing.T) {
	flux.Just().Subscribe(
		context.Background(),
		rs.OnNext(func(v interface{}) {
			log.Println("next:", v)
		}),
		rs.OnComplete(func() {
			log.Println("complete")
		}),
	)
}
