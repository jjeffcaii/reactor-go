### Just
Use `Just` to wrap multi existing elements.

``` go
package main

import "github.com/jjeffcaii/reactor-go/flux"

func main() {
    // wrap some elements
	flux.Just("foo", "bar", "qux")
    // wrap a slice
    flux.Just([]interface{}{"foo", "bar", "qux"}...)
}
```

### Create
Use `Create` to create a new `Flux`.<br/>
Flux is **LAZY** and nothing happends unless you **SUBSCRIBE** it!

``` go
package main

import (
	"context"

	"github.com/jjeffcaii/reactor-go/flux"
)

func main() {
	flux.Create(func(ctx context.Context, sink flux.Sink) {
		// produce 3 elements
		sink.Next("foo")
		sink.Next("bar")
		sink.Next("qux")
		// mark complete with success
		sink.Complete()
		// or emit an error.
		//sink.Error(errors.New("something wrong"))
	})
}
```

### Interval
Use `Interval` to create a `Flux`.<br/>

``` go
package main

import (
	"context"
	"fmt"
	"time"

	"github.com/jjeffcaii/reactor-go/flux"
)

func main() {

	// IntervalFlux will be scheduled async by default, this is the reason we use need block last element.
	// And It will emit infinite int64 sequences.
    // We take 4 elements in this example.

	_, _ = flux.Interval(500 * time.Millisecond).
    	Take(4).
		DoOnNext(func(value interface{}) {
			// value is an int64
			fmt.Println("next:", value)
		}).
		BlockLast(context.Background())

	// It will print some numbers:
	// next: 0
	// next: 1
	// next: 2
	// next: 3
}
```

### Range
Use `Range` to create a `Flux` with some consecutive numbers.

``` go
package main

import (
	"context"
	"fmt"

	"github.com/jjeffcaii/reactor-go/flux"
)

func main() {
	flux.
		Range(0, 3).
		Map(func(v interface{}) interface{} {
			return fmt.Sprintf("Element %d", v.(int))
		}).
		DoOnNext(func(v interface{}) {
			fmt.Println("next:", v)
		}).
		Subscribe(context.Background())

	// Should print:
	// next: Element 0
	// next: Element 1
	// next: Element 2
}
```

### Empty
Use `Empty` to create a `Flux` without any elements.

### Error
Use `Error` to wrap an error to a `Flux`.
