package mono_test

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/jjeffcaii/reactor-go"
	"github.com/jjeffcaii/reactor-go/mono"
	"github.com/stretchr/testify/assert"
)

func TestSwitchIfError(t *testing.T) {
	fakeErr := errors.New("fake error")

	// should call switch to another Mono if error occurs
	value, err := mono.Error(fakeErr).
		Map(func(any reactor.Any) (reactor.Any, error) {
			assert.FailNow(t, "should be unreachable")
			return nil, nil
		}).
		SwitchIfError(func(err error) mono.Mono {
			assert.Equal(t, fakeErr, err)
			return mono.Just(1)
		}).
		Map(func(any reactor.Any) (reactor.Any, error) {
			return any.(int) * 2, nil
		}).
		Block(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, 2, value)

	// should not call SwitchIfError when Mono ends with a successful value
	value, err = mono.Just(1).
		Map(func(any reactor.Any) (reactor.Any, error) {
			return any.(int) + 1, nil
		}).
		SwitchIfError(func(err error) mono.Mono {
			assert.FailNow(t, "should be unreachable")
			return mono.Error(err)
		}).
		Map(func(any reactor.Any) (reactor.Any, error) {
			return any.(int) * 2, nil
		}).
		Block(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, 4, value)

	// nothing happened if empty
	value, err = mono.Empty().
		SwitchIfError(func(err error) mono.Mono {
			assert.FailNow(t, "should be unreachable")
			return mono.Error(err)
		}).
		Map(func(any reactor.Any) (reactor.Any, error) {
			assert.FailNow(t, "should be unreachable")
			return nil, nil
		}).
		Block(context.Background())
	assert.NoError(t, err)
	assert.Nil(t, value)

	// check oneshot api
	value, err = mono.ErrorOneshot(fakeErr).
		SwitchIfError(func(err error) mono.Mono {
			assert.Equal(t, fakeErr, err)
			return mono.Just(1)
		}).
		Block(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, 1, value)

	delay := 500 * time.Millisecond

	done := make(chan struct{})
	var end time.Time
	mono.Error(fakeErr).
		SwitchIfError(func(err error) mono.Mono {
			// simulate async mono
			return mono.Delay(delay).
				Map(func(any reactor.Any) (reactor.Any, error) {
					return 1, nil
				}).
				DoFinally(func(s reactor.SignalType) {
					close(done)
				})
		}).
		DoOnNext(func(v reactor.Any) error {
			assert.Equal(t, 1, v)
			return nil
		}).
		DoOnError(func(e error) {
			assert.FailNow(t, "should be unreachable")
		}).
		DoOnComplete(func() {
			end = time.Now()
		}).
		Subscribe(context.Background())
	start := time.Now()
	fmt.Println("should not block here, done will be finished after", delay)
	<-done
	assert.True(t, int64(end.Sub(start)) > int64(delay))
}
