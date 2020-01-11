# Map
Use `Map` to transform origin element.

``` go
package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/jjeffcaii/reactor-go/flux"
)

func main() {
	flux.
		Just("foo", "bar", "qux").
		Map(func(v interface{}) interface{} {
			// Transform origin element here.
			return fmt.Sprintf("Hello %s!", strings.Title(v.(string)))
		}).
		DoOnNext(func(v interface{}) {
			fmt.Println("next:", v)
		}).
		Subscribe(context.Background())
	// Should print:
	//next: Hello Foo!
	//next: Hello Bar!
	//next: Hello Qux!
}

```
