### SwitchOnFirst
Get first element in this `Flux` and transform to another `Flux`.

``` go
package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/jjeffcaii/reactor-go/flux"
)

func main() {
	flux.Just("golang", "I love golang.", "I love java.", "Awesome golang.", "I love ruby.").
		SwitchOnFirst(func(firstSignal flux.Signal, originFlux flux.Flux) flux.Flux {
			// extract first word and filtering word that contains it.
			first, ok := firstSignal.Value()
			if !ok {
				return originFlux
			}
			return originFlux.
				Filter(func(item interface{}) bool {
					return strings.Contains(item.(string), first.(string))
				})
		}).
		DoOnNext(func(v interface{}) {
			fmt.Println("next:", v)
		}).
		Subscribe(context.Background())
	// Should print:
	// next: golang
	// next: I love golang.
	// next: Awesome golang.
}

```
