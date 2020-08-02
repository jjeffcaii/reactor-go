# DoFinally
Add behavior triggered after the `Mono`/`Flux` terminates for any reason, including cancellation.

``` go
package main

import (
	"context"
	"errors"
	"fmt"

	"github.com/jjeffcaii/reactor-go"
	"github.com/jjeffcaii/reactor-go/flux"
	"github.com/jjeffcaii/reactor-go/mono"
)

func main() {
	mono.Just(42).
		DoFinally(func(s rs.SignalType) {
			fmt.Println("trigger finally")
		}).
		DoOnNext(func(v interface{}) {
			fmt.Println("next:", v)
		}).
		DoOnComplete(func() {
			fmt.Println("completed")
		}).
		Subscribe(context.Background())
	// Should print:
	// next: 42
	// completed
	// trigger finally

	flux.
		Error(errors.New("something wrong")).
		DoFinally(func(s rs.SignalType) {
			fmt.Println("trigger finally")
		}).
		DoOnError(func(e error) {
			fmt.Println("got error:", e)
		}).
		Subscribe(context.Background())
	// Should print:
	// got error: something wrong
	// trigger finally
}
```
