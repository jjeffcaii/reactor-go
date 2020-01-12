### SwitchIfEmpty
Fallback to an alternative `Mono` if this mono is completed without data.

``` go
package main

import (
	"context"
	"fmt"

	"github.com/jjeffcaii/reactor-go/mono"
)

func main() {
	// v1 should be 42
	v1, _ := mono.Empty().SwitchIfEmpty(mono.Just(42)).Block(context.Background())
	fmt.Println("v1:", v1)

	// v2 should be 1024
	v2, _ := mono.Just(1024).SwitchIfEmpty(mono.Just(2048)).Block(context.Background())
	fmt.Println("v2:", v2)
}

```
