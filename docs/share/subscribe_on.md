# SubscribeOn
Specifying a scheduler when subscribing a `Mono`/`Flux`.</br>

There are many schedulers in `scheduler` package:
 - Elastic: a scheduler within a goroutine pool.
 - Single: a scheduler which contains only one worker.
 - Parallel: a scheduler with a fixed number of goroutine.
 - Immediate: a scheduler which executes tasks on the caller's goroutine immediately.

``` go
package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/jjeffcaii/reactor-go/mono"
	"github.com/jjeffcaii/reactor-go/scheduler"
)

func main() {
    // Example for request a REST API async and block for result.
	mono.
		Create(func(ctx context.Context, s mono.Sink) {
            // Start a request with Context API.
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
		SubscribeOn(scheduler.Elastic()). // <-- use default elastic scheduler here.
		Block(context.Background()) // <-- block for result.
	// next: x.x.x.x
}

```
