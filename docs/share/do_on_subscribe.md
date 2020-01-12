## DoOnSubscribe
Add behavior triggered when the `Mono`/`Flux` is done being subscribed.

``` go
package main

import (
	"context"
	"fmt"

	"github.com/jjeffcaii/reactor-go"
	"github.com/jjeffcaii/reactor-go/mono"
)

func main() {
	mono.
		Create(func(ctx context.Context, s mono.Sink) {
			s.Success(42)
		}).
		DoOnSubscribe(func(su rs.Subscription) {
			fmt.Println("subscribed!")
		}).
		DoOnNext(func(v interface{}) {
			fmt.Println("next:", v)
		}).
		Subscribe(context.Background())
	// Should print:
	// subscribed!
	// next: 42
}
```
