### Zip
Use `Zip` to wrap an existing element.

```go
package main

import (
	"context"
	"fmt"
	"github.com/jjeffcaii/reactor-go"
	"github.com/jjeffcaii/reactor-go/mono"
	"github.com/jjeffcaii/reactor-go/scheduler"
	"testing"
)

func TestZipMono(t *testing.T) {
	mono.Zip(mono.JustOneshot(1).Map(func(any reactor.Any) (reactor.Any, error) {
		return any.(int)+1,nil
	}).SubscribeOn(scheduler.Parallel()), mono.JustOneshot(3).Map(func(any reactor.Any) (reactor.Any, error) {
		return any.(int)+1,nil
	}).SubscribeOn(scheduler.Parallel())).Map(func(any reactor.Any) (reactor.Any, error) {
		var a = any.(*mono.Items)
		fmt.Println("-----",a.It[1],a.It[0])
		return any,nil
	}).Subscribe(context.Background())
	mono.Zip(mono.Error(fmt.Errorf("ddddddddd")),mono.Empty()).Map(func(any reactor.Any) (reactor.Any, error) {
		var a = any.(*mono.Items)
		fmt.Println("-----",a.It[1],a.It[0])
		return any,nil
	}).DoOnError(func(e error) {
		fmt.Println("+++++")
	}).Subscribe(context.Background())
}

```