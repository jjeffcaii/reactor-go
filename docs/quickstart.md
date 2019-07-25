# Quick Start

## Installation
It is recommended to install `reactor-go` as a [Go module](https://github.com/golang/go/wiki/Modules).

``` shell
$ go get -u github.com/jjeffcaii/reactor-go
$ go get -u github.com/jjeffcaii/reactor-go@vX.Y.Z
```

## Imports
We offer `mono` and `flux` package for most features. First you need import these packages.

``` go
import (
  "github.com/jjeffcaii/reactor-go"
  "github.com/jjeffcaii/reactor-go/mono"
  "github.com/jjeffcaii/reactor-go/flux"
)
```

## Mono
Here's a tiny example which show how mono API works. It'll print a `Hello World!` string.

``` go
package main

import (
	"context"
	"fmt"

	"github.com/jjeffcaii/reactor-go/mono"
)

func main() {
	mono.Just("World").
		Map(func(v interface{}) interface{} {
			return fmt.Sprintf("Hello %s", v)
		}).
		DoOnNext(func(v interface{}) {
			fmt.Println(v)
		}).
		Subscribe(context.Background())
}
```

## Flux
Here's a tiny example which show how flux API works. It'll print `Hello Tom!`, `Hello Jerry!` and `Hello Tux!`.

``` go
package main

import (
	"context"
	"fmt"

	"github.com/jjeffcaii/reactor-go/flux"
)

func main() {
	flux.Just("Tom", "Jerry", "Tux").
		Map(func(v interface{}) interface{} {
			return fmt.Sprintf("Hello %d!", v.(int))
		}).
		DoOnNext(func(v interface{}) {
			fmt.Println(v)
		}).
		Subscribe(context.Background())
}

```
