package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/jjeffcaii/reactor-go/flux"
	"github.com/jjeffcaii/reactor-go/mono"
	"github.com/jjeffcaii/reactor-go/scheduler"
)

func main() {

	// Interval Flux will schedule async as default. (This is the reason we use `BlockLast`.)
	// And It will emit infinite int64 sequences.
	_, _ = flux.Interval(500 * time.Millisecond).
		Take(4).
		DoOnNext(func(value interface{}) {
			// value is an int64
			fmt.Println("next:", value)
		}).
		BlockLast(context.Background())
	// It will print some numbers:
	// next: 0
	// next: 1
	// next: 2
	// next: 3

	mono.
		Create(func(ctx context.Context, s mono.Sink) {
			req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://ipecho.net/plain", nil)
			if err != nil {
				s.Error(err)
				return
			}
			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				s.Error(err)
				return
			}
			defer resp.Body.Close()
			if bs, err := ioutil.ReadAll(resp.Body); err == nil {
				s.Success(string(bs))
			} else {
				s.Error(err)
			}

		}).
		DoOnNext(func(v interface{}) {
			fmt.Println("next:", v)
		}).
		DoOnError(func(e error) {
			fmt.Println("oops:", e)
		}).
		SubscribeOn(scheduler.Elastic()).
		Block(context.Background())
	// next: x.x.x.x
}
