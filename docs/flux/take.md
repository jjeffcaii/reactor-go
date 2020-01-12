### Take
Take only the first N values from this `Flux`, if available.

``` go
package main

import (
	"context"
	"fmt"
	"time"

	"github.com/jjeffcaii/reactor-go/flux"
)

func main() {
	// Example: taking 3 elements
	flux.
		Interval(500 * time.Millisecond).
		Take(3).
		DoOnNext(func(v interface{}) {
			fmt.Println(time.Now().Format("2006-01-02 03:04:05"), "next:", v)
		}).
		BlockLast(context.Background())
	// Should print:
	// 2020-01-12 07:26:58 next: 0
	// 2020-01-12 07:26:59 next: 1
	// 2020-01-12 07:26:59 next: 2
}

```
