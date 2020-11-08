### Block
Subscribe to this Mono and block indefinitely until a next signal is received.

```go
package main

import (
	"context"
	"fmt"

	"github.com/jjeffcaii/reactor-go/mono"
)

func main() {
	value, err := mono.Just(42).Block(context.Background())
	if err != nil {
		panic(err)
	}
	fmt.Println("next:", value)
	// Should print "next: 42"

	value, err = mono.Empty().Block(context.Background())
	if err != nil {
		panic(err)
	}
	fmt.Println("next:", value)
	// Should print "next: <nil>"
}

```