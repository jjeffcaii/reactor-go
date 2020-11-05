package reactor_test

import (
	"context"
	"fmt"
	"github.com/jjeffcaii/reactor-go"
	"github.com/jjeffcaii/reactor-go/mono"
	"testing"
)


func TestSwitchIfEmpty(t *testing.T) {
	v, err := mono.Empty().SwitchIfEmpty(mono.Just(333 * 2)).Block(context.Background())
	fmt.Println("-------",v,err)
	vv, e := mono.JustOneshot(15555).Map(func(any reactor.Any) (reactor.Any, error) {
		fmt.Println("+++++++++",any)
		return any,nil
	}).Map(func(any reactor.Any) (reactor.Any, error) {
		fmt.Println("+++++++++1111",any)
		return any,fmt.Errorf("cccccccc")
	}).DoOnError(func(e error) {
		fmt.Println("---=====-----",e.Error())
	}).Map(func(any reactor.Any) (reactor.Any, error) {
		fmt.Println("11111+++++++++1111",any)
		return any,nil
	}).CreatteMonoIfError(123).
		Map(func(any reactor.Any) (reactor.Any, error) {
		fmt.Println("323234+++++++++1111",any)
		return any.(int)+100,nil
	}).Block(context.Background())
	fmt.Println("-------",vv,e)
}