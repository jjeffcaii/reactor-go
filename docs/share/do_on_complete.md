# DoOnComplete
Add behavior triggered when the `Mono`/`Flux` completes successfully.

``` go
package main

import (
	"context"
	"errors"
	"fmt"

	"github.com/jjeffcaii/reactor-go/mono"
)

func main() {
	mono.Just(42).
		DoOnComplete(func() {
			fmt.Println("completed!")
		}).
		Subscribe(context.Background())
	// completed!
	mono.Error(errors.New("something wrong")).
		DoOnError(func(e error) {
			fmt.Println("got error:", e)
		}).
		Subscribe(context.Background())
	// got error: something wrong
}
```
