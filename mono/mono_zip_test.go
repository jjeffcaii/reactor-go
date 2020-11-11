package mono_test

import (
	"context"
	"fmt"
	"github.com/jjeffcaii/reactor-go"
	"github.com/jjeffcaii/reactor-go/mono"
	"github.com/jjeffcaii/reactor-go/scheduler"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestZipMono(t *testing.T) {
	mono.Zip(mono.JustOneshot(1).Map(func(any reactor.Any) (reactor.Any, error) {
		return any.(int)+1,nil
	}).SubscribeOn(scheduler.Parallel()), mono.JustOneshot(3).Map(func(any reactor.Any) (reactor.Any, error) {
		return any.(int)+1,nil
	}).SubscribeOn(scheduler.Parallel())).Map(func(any reactor.Any) (reactor.Any, error) {
		var a = any.(*mono.Items)
		t.Log("-----",a.It[1],a.It[0])
		assert.NoError(t, a.It[1].Err, "should not return error")
		assert.NoError(t, a.It[0].Err, "should not return error")
		return any,nil
	}).Subscribe(context.Background())
	var r,e = mono.Zip(mono.Error(fmt.Errorf("ddddddddd")),mono.Empty()).Map(func(any reactor.Any) (reactor.Any, error) {
		var a = any.(*mono.Items)
		t.Log("-----",a.It[1],a.It[0])
		return any,nil
	}).DoOnError(func(e error) {
		t.Log("DoOnError")
	}).Block(context.Background())
	t.Log("mono.zip.block:",r)
	assert.NoError(t, e, "should not return error")
}