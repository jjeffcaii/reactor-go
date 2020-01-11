# DoOnNext
Add behavior triggered when the `Mono`/`Flux` emits an element.

``` go
package main

import (
	"context"
	"fmt"

	"github.com/jjeffcaii/reactor-go/flux"
	"github.com/jjeffcaii/reactor-go/mono"
)

func main() {
	// DoOnNext example for Mono
	mono.Just(42).
		DoOnNext(func(v interface{}) {
			fmt.Println("next:", v)
		}).
		Subscribe(context.Background())
	// DoOnNext example for Flux
	flux.Just(1, 2, 3).
		DoOnNext(func(v interface{}) {
			fmt.Println("next:", v)
		}).
		Subscribe(context.Background())
}
```
