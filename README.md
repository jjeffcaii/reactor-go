<h1 align="center">Welcome to reactor-go 👋</h1>

[![Build Status](https://travis-ci.com/jjeffcaii/reactor-go.svg?branch=master)](https://travis-ci.com/jjeffcaii/reactor-go)
![GoDoc](https://godoc.org/github.com/jjeffcaii/reactor-go?status.svg)
[![Go Report Card](https://goreportcard.com/badge/github.com/jjeffcaii/reactor-go)](https://goreportcard.com/report/github.com/jjeffcaii/reactor-go)
![License](https://img.shields.io/github/license/jjeffcaii/reactor-go.svg)
![GitHub Release](https://img.shields.io/github/release-pre/jjeffcaii/reactor-go.svg)

> A golang implementation for reactive-streams. ***[WARNNING] IT IS STILL UNDER DEVELOPMENT!!!***

### 🏠 [Homepage](https://github.com/jjeffcaii/reactor-go)

## Install

```sh
go get -u github.com/jjeffcaii/reactor-go
```

## Example

### Mono
```go
package mono_test

import (
	"context"
	"fmt"

	rs "github.com/jjeffcaii/reactor-go"
	"github.com/jjeffcaii/reactor-go/mono"
)

func Example() {
	gen := func(ctx context.Context, sink mono.Sink) {
		sink.Success("World")
	}
	mono.
		Create(gen).
		Map(func(i interface{}) interface{} {
			return "Hello " + i.(string) + "!"
		}).
		Subscribe(context.Background(), rs.NewSubscriber(
			rs.OnNext(func(s rs.Subscription, v interface{}) {
				fmt.Println(v)
			}),
		))
}

// Should print
// Hello World!

```

### Flux
```go
package rs_test

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
	flux.Create(gen).
		Filter(func(i interface{}) bool {
			return i.(int)%2 == 0
		}).
		Map(func(i interface{}) interface{} {
			return fmt.Sprintf("#HELLO_%04d", i.(int))
		}).
		SubscribeOn(scheduler.Elastic()).
		Subscribe(context.Background(), rs.NewSubscriber(
			rs.OnSubscribe(func(s rs.Subscription) {
				s.Request(1)
			}),
			rs.OnNext(func(s rs.Subscription, v interface{}) {
				fmt.Println("next:", v)
				s.Request(1)
			}),
			rs.OnComplete(func() {
				close(done)
			}),
		))
	<-done
}
// Should print:
// next: #HELLO_0000
// next: #HELLO_0002
// next: #HELLO_0004
// next: #HELLO_0006
// next: #HELLO_0008
```

## Author

👤 **Jeffsky**

* Github: [@jjeffcaii](https://github.com/jjeffcaii)

## Show your support

Give a ⭐️ if this project helped you!

***
_This README was generated with ❤️ by [readme-md-generator](https://github.com/kefranabg/readme-md-generator)_
