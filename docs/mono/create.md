### Just
Use `Just` to wrap an existing element.

``` go
package main

import "github.com/jjeffcaii/reactor-go/mono"

func main() {
	mono.Just("Hello World!")
}
```

### Create
Use `Create` to create a new `Mono`.<br/>
Mono is **LAZY** and nothing happens unless you call `Subscribe`!

``` go
package main

import (
	"context"
	"errors"

	"github.com/jjeffcaii/reactor-go/mono"
)

func main() {
	mono.Create(func(ctx context.Context, s mono.Sink) {
		// Do something and call Success.
		s.Success("Hello World!")
	})
	mono.Create(func(ctx context.Context, s mono.Sink) {
		// Do something but failed.
		s.Error(errors.New("something bad"))
	})
}
```

### Delay
Create a Mono which delays an `int64` by a given duration. A delayed Mono will be scheduled asynchronously.<br/>
In example below, we call `Block` to block result in current coroutine. So it should print `Bingo!!!` after 3 seconds.

``` go
package main

import (
	"context"
	"fmt"
	"time"

	"github.com/jjeffcaii/reactor-go/mono"
)

func main() {
	mono.Delay(3 * time.Second).
		DoOnNext(func(v interface{}) {
			fmt.Println("Bingo!!!")
		}).
		Block(context.Background())
}
```

### Empty
Use `Empty` to create a `Mono` without any element.

### Error
Use `Error` to wrap an error to a `Mono`.
