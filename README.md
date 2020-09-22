# reactor-go 🚀🚀🚀

[![Build Status](https://travis-ci.com/jjeffcaii/reactor-go.svg?branch=master)](https://travis-ci.com/jjeffcaii/reactor-go)
[![codecov](https://codecov.io/gh/jjeffcaii/reactor-go/branch/master/graph/badge.svg)](https://codecov.io/gh/jjeffcaii/reactor-go)
[![GoDoc](https://godoc.org/github.com/jjeffcaii/reactor-go?status.svg)](https://godoc.org/github.com/jjeffcaii/reactor-go)
[![Go Report Card](https://goreportcard.com/badge/github.com/jjeffcaii/reactor-go)](https://goreportcard.com/report/github.com/jjeffcaii/reactor-go)
[![License](https://img.shields.io/github/license/jjeffcaii/reactor-go.svg)](https://github.com/jjeffcaii/reactor-go/blob/master/LICENSE)
[![GitHub Release](https://img.shields.io/github/release-pre/jjeffcaii/reactor-go.svg)](https://github.com/jjeffcaii/reactor-go/releases)

> A golang implementation for [reactive-streams](https://www.reactive-streams.org/).
<br>🚧🚧🚧 ***IT IS UNDER ACTIVE DEVELOPMENT!!!***
<br>⚠️⚠️⚠️ ***DO NOT USE IN ANY PRODUCTION ENVIRONMENT!!!***

## Install

```sh
go get -u github.com/jjeffcaii/reactor-go
```

## Example

> NOTICE:
<br> We can only use `func(interface{})interface{}` for most operations because Golang has not Generics. 😭
<br> If you have any better idea, please let me know. 😀

### Mono
```go
package mono_test

import (
	"context"
	"fmt"

	"github.com/jjeffcaii/reactor-go"
	"github.com/jjeffcaii/reactor-go/mono"
)

func Example() {
	gen := func(ctx context.Context, sink mono.Sink) {
		sink.Success("World")
	}
	mono.
		Create(gen).
		Map(func(input reactor.Any) (output reactor.Any, err error) {
			output = "Hello " + input.(string) + "!"
			return
		}).
		DoOnNext(func(v reactor.Any) error {
			fmt.Println(v)
			return nil
		}).
		Subscribe(context.Background())
}

// Should print
// Hello World!

```

### Flux
```go
package flux_test

import (
	"context"
	"fmt"

	"github.com/jjeffcaii/reactor-go"
	"github.com/jjeffcaii/reactor-go/flux"
	"github.com/jjeffcaii/reactor-go/scheduler"
)

func Example() {
	gen := func(ctx context.Context, sink flux.Sink) {
		for i := 0; i < 10; i++ {
			v := i
			sink.Next(v)
		}
		sink.Complete()
	}
	done := make(chan struct{})

	var su reactor.Subscription
	flux.Create(gen).
		Filter(func(i interface{}) bool {
			return i.(int)%2 == 0
		}).
		Map(func(input reactor.Any) (output reactor.Any, err error) {
			output = fmt.Sprintf("#HELLO_%04d", input.(int))
			return
		}).
		SubscribeOn(scheduler.Elastic()).
		Subscribe(context.Background(),
			reactor.OnSubscribe(func(s reactor.Subscription) {
				su = s
				s.Request(1)
			}),
			reactor.OnNext(func(v reactor.Any) error {
				fmt.Println("next:", v)
				su.Request(1)
				return nil
			}),
			reactor.OnComplete(func() {
				close(done)
			}),
		)
	<-done
}
// Should print:
// next: #HELLO_0000
// next: #HELLO_0002
// next: #HELLO_0004
// next: #HELLO_0006
// next: #HELLO_0008
```
