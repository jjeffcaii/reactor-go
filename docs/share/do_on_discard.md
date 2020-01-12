## DoOnDiscard
Add behavior triggered when filtering/dropping elements in `Mono`/`Flux`.

``` go
package main

import (
	"context"
	"fmt"

	"github.com/jjeffcaii/reactor-go/flux"
)

func main() {
	flux.
		Range(0, 5).
		Filter(func(v interface{}) bool {
			return v.(int)%2 == 0
		}).
		DoOnDiscard(func(v interface{}) {
			fmt.Println("discard:", v)
		}).
		Subscribe(context.Background())
	// Should print:
	// discard: 1
	// discard: 3
}
```
