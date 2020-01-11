# Filter
You can easily filter out some elements.

``` go
package main

import (
	"context"
	"fmt"

	"github.com/jjeffcaii/reactor-go/flux"
)

func main() {
	flux.
		Range(0, 10).
		Filter(func(v interface{}) bool {
			return v.(int)%2 == 0
		}).
		DoOnNext(func(v interface{}) {
			fmt.Println("next:", v)
		}).
		Subscribe(context.Background())

	// Should print:
	// next: 0
	// next: 2
	// next: 4
	// next: 6
	// next: 8
}
```
