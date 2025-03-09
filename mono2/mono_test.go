package mono_test

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/jjeffcaii/reactor-go"
	. "github.com/jjeffcaii/reactor-go/mono2"
)

var fakeErr = errors.New("fake error")

func TestMono(t *testing.T) {
	var actual int

	Just(10).
		DoOnNext(func(ctx context.Context, next int) error {
			t.Log("DoOnNext:", next)
			return nil
		}).
		Filter(func(ctx context.Context, next int) bool {
			return next%2 == 0
		}).
		Map(func(ctx context.Context, next int) (int, error) {
			return next * 10, nil
		}).
		Subscribe(context.TODO(), reactor.OnNext(func(ctx context.Context, next int) error {
			actual = next
			t.Log("Subscribe:", next)
			return nil
		}))

	assert.Equal(t, 100, actual)
}

func TestEmpty(t *testing.T) {
	v, _ := Empty[int]().SwitchValueIfEmpty(1234).Block(context.TODO())
	assert.Equal(t, 1234, v.V)
}

func TestError(t *testing.T) {
	v, err := Error[int](fakeErr).SwitchValueIfError(123).Block(context.TODO())
	assert.NoError(t, err)
	assert.True(t, v.Valid)
	assert.Equal(t, 123, v.V)
}

func TestMonoBlock(t *testing.T) {
	v, err := Just(10).Block(context.TODO())
	assert.NoError(t, err, "block should not return error")
	assert.True(t, v.Valid, "block value should be valid")
	assert.Equal(t, 10, v.V, "block value should be 10")

	v, err = Just(10).
		Filter(func(ctx context.Context, next int) bool {
			return next > 100
		}).
		Block(context.TODO())
	assert.NoError(t, err, "block should not return error")
	assert.False(t, v.Valid, "block value should be invalid")
	assert.Equal(t, 0, v.V, "block should return zero value")
}

func TestMap(t *testing.T) {
	now := Create[time.Time](func(ctx context.Context, sink Sink[time.Time]) {
		sink.Success(time.Now())
	})

	addDay := now.Map(func(ctx context.Context, next time.Time) (time.Time, error) {
		return next.AddDate(0, 0, 1), nil
	})

	toUnixNano := Map[time.Time, int64](now, func(ctx context.Context, next time.Time) (int64, error) {
		return next.UnixNano(), nil
	})

	a, _ := now.Block(context.TODO())
	t.Log("now:", a)

	time.Sleep(1 * time.Second)
	b, _ := addDay.Block(context.TODO())
	t.Log("addDay:", b)

	time.Sleep(1 * time.Second)
	c, _ := toUnixNano.Block(context.TODO())
	t.Log("toUnixNano:", c)

}

func TestFlatMap(t *testing.T) {
	fm := FlatMap[int, string](Just[int](1), func(ctx context.Context, value int) Mono[string] {
		return Create[string](func(ctx context.Context, sink Sink[string]) {
			resp, err := http.Get(fmt.Sprintf("https://httpbin.org/get?v=%d", value))
			if err != nil {
				sink.Error(err)
				return
			}
			defer resp.Body.Close()

			b, err := io.ReadAll(resp.Body)
			if err != nil {
				sink.Error(err)
				return
			}

			sink.Success(string(b))
		})
	})

	v, err := fm.Block(context.TODO())
	assert.NoError(t, err, "block should not return error")
	assert.True(t, v.Valid, "block value should be valid")
	t.Log("v:", v)

}
