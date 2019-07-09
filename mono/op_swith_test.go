package mono_test

import (
	"context"
	"log"
	"testing"

	"github.com/jjeffcaii/reactor-go/mono"
	"github.com/stretchr/testify/assert"
)

func TestSwitchIfEmpty(t *testing.T) {
	mono.Empty().
		SwitchIfEmpty(mono.Just(333)).
		Map(func(i interface{}) interface{} {
			return i.(int) * 2
		}).
		DoOnNext(func(v interface{}) {
			assert.Equal(t, 666, v.(int), "bad result")
			log.Println("next:", 666)
		}).
		Subscribe(context.Background())

}
