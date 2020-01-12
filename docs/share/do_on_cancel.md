## DoOnCancel
Add behavior triggered when the `Mono`/`Flux` is cancelled.

``` go
package main

import (
	"context"
	"fmt"

	rs "github.com/jjeffcaii/reactor-go"
	"github.com/jjeffcaii/reactor-go/mono"
)

func main() {
	mono.
		Create(func(ctx context.Context, s mono.Sink) {
			s.Success(42)
		}).
		DoOnCancel(func() {
            // Do something when Mono cancelled
			fmt.Println("cancelled")
		}).
		Subscribe(context.Background(), rs.OnSubscribe(func(su rs.Subscription) {
            // cancel here
			su.Cancel()
		}))
	// Should print:
	// cancelled
}

```
