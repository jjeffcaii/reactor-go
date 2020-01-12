## DelayElement
Delay each of this `Mono`/`Flux` elements by a given Duration.

``` go
package main

import (
	"context"
	"fmt"
	"time"

	"github.com/jjeffcaii/reactor-go/flux"
)

func main() {
	flux.
		Range(0, 5).
		DelayElement(1 * time.Second).
		DoOnNext(func(v interface{}) {
			fmt.Println(time.Now().Format("2006-01-02 03:04:05"), "next:", v)
		}).
		DoOnComplete(func() {
			fmt.Println("completed")
		}).
		BlockLast(context.Background())
	// Should print:
	// 2020-01-12 05:41:00 next: 0
	// 2020-01-12 05:41:01 next: 1
	// 2020-01-12 05:41:02 next: 2
	// 2020-01-12 05:41:03 next: 3
	// 2020-01-12 05:41:04 next: 4
	// completed
}

```
