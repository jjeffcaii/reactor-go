### BlockLast
Subscribe to the Flux and block indefinitely until the upstream signals its last value or error.

```go
package main

import (
	"context"
	"log"
	"time"

	"github.com/jjeffcaii/reactor-go/flux"
)

func main() {
	value, err := flux.Interval(1 * time.Second).Take(3).BlockLast(context.Background())
	if err != nil {
		panic(err)
	}
	log.Println("last:", value)
	// Should print "last: 2" after 3 seconds.
}

```
