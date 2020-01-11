# DoOnError
Add behavior triggered when the `Mono`/`Flux` completes with an error.

``` go
package main

import (
	"context"
	"errors"
	"fmt"

	"github.com/jjeffcaii/reactor-go/mono"
)

func main() {
	// Create an Error Mono.
	mono.Error(errors.New("something wrong")).
		DoOnError(func(e error) {
			fmt.Println("got error:", e)
		}).
		Subscribe(context.Background())
	// got error: something wrong

	// Create a simple Mono and panic an error in DoOnNext.
	mono.Just(1).
		DoOnNext(func(v interface{}) {
			panic(errors.New("bad thing"))
		}).
		DoOnError(func(e error) {
			fmt.Println("got error:", e)
		}).
		Subscribe(context.Background())
    // got error: bad thing
}
```
