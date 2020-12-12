package mono_test

import (
	"context"
	"testing"

	"github.com/jjeffcaii/reactor-go"
	"github.com/jjeffcaii/reactor-go/mono"
	"github.com/stretchr/testify/assert"
)

const (
	cntDoOnNext int = iota
	cntDoOnError
	cntDoOnComplete
	cntOnNext
	cntOnError
	cntOnComplete
	cntTotal
)

func TestPanic_Map(t *testing.T) {
	var cnt [cntTotal]int
	runPanicTest(t, func(m mono.Mono) mono.Mono {
		return m.Map(func(any reactor.Any) (reactor.Any, error) {
			panic("fake panic")
		})
	}, cnt[:])
	normalPanicCheck(t, cnt[:])
}

func TestPanic_Filter(t *testing.T) {
	var cnt [cntTotal]int
	runPanicTest(t, func(m mono.Mono) mono.Mono {
		return m.Filter(func(any reactor.Any) bool {
			panic("fake panic")
		})
	}, cnt[:])
	normalPanicCheck(t, cnt[:])
}

func TestPanic_DoOnNext(t *testing.T) {
	var cnt [cntTotal]int
	runPanicTest(t, func(m mono.Mono) mono.Mono {
		return m.DoOnNext(func(v reactor.Any) error {
			panic("fake panic")
		})
	}, cnt[:])
	normalPanicCheck(t, cnt[:])
}

func TestPanic_FlatMap(t *testing.T) {
	var cnt [cntTotal]int
	runPanicTest(t, func(m mono.Mono) mono.Mono {
		return m.FlatMap(func(any reactor.Any) mono.Mono {
			panic("fake panic")
		})
	}, cnt[:])
	normalPanicCheck(t, cnt[:])
	for i := 0; i < len(cnt); i++ {
		cnt[i] = 0
	}
	runPanicTest(t, func(m mono.Mono) mono.Mono {
		return m.FlatMap(func(any reactor.Any) mono.Mono {
			return nil
		})
	}, cnt[:])
	normalPanicCheck(t, cnt[:])
}

func TestPanic_SwitchIfError(t *testing.T) {
	run := func(sw func(err error) mono.Mono) {
		var cnt [cntTotal]int
		var one int
		mono.Error(fakeErr).
			DoOnError(func(e error) {
				one++
			}).
			SwitchIfError(sw).
			DoOnNext(func(v reactor.Any) error {
				cnt[cntDoOnNext]++
				return nil
			}).
			DoOnError(func(e error) {
				cnt[cntDoOnError]++
			}).
			Subscribe(context.Background(), reactor.OnError(func(e error) {
				cnt[cntOnError]++
			}), reactor.OnNext(func(v reactor.Any) error {
				cnt[cntOnNext]++
				return nil
			}), reactor.OnComplete(func() {
				cnt[cntOnComplete]++
			}))
		normalPanicCheck(t, cnt[:])
	}

	run(func(err error) mono.Mono {
		panic("fake panic")
	})

	run(func(err error) mono.Mono {
		return nil
	})

	run(nil)

	run(func(err error) mono.Mono {
		return mono.Just(1).Map(func(any reactor.Any) (reactor.Any, error) {
			panic("fake panic")
		})
	})

}

func TestPanic_SwitchIfEmpty(t *testing.T) {
	run := func(replace mono.Mono) {
		var zero int
		var cnt [cntTotal]int
		mono.Empty().
			DoOnNext(func(v reactor.Any) error {
				zero++
				return nil
			}).
			DoOnError(func(e error) {
				zero++
			}).
			SwitchIfEmpty(replace).
			DoOnNext(func(v reactor.Any) error {
				cnt[cntDoOnNext]++
				return nil
			}).
			DoOnError(func(e error) {
				cnt[cntDoOnError]++
				t.Log("[PANIC] DoOnError:", e)
			}).
			Subscribe(context.Background(),
				reactor.OnNext(func(v reactor.Any) error {
					cnt[cntOnNext]++
					return nil
				}),
				reactor.OnError(func(e error) {
					cnt[cntOnError]++
					t.Log("[PANIC] DoOnError:", e)
				}),
				reactor.OnComplete(func() {
					cnt[cntOnComplete]++
				}),
			)
		assert.Zero(t, zero)
		normalPanicCheck(t, cnt[:])
	}

	run(mono.Just(1).
		Map(func(any reactor.Any) (reactor.Any, error) {
			panic("fake panic")
		}))

	run(nil)
}

func normalPanicCheck(t *testing.T, cnt []int) {
	assert.Equal(t, 0, cnt[cntDoOnNext])
	assert.Equal(t, 1, cnt[cntDoOnError])
	assert.Equal(t, 0, cnt[cntDoOnComplete])
	assert.Equal(t, 1, cnt[cntOnError])
	assert.Equal(t, 0, cnt[cntOnNext])
	assert.Equal(t, 0, cnt[cntOnComplete])
}

func runPanicTest(t *testing.T, convert func(mono.Mono) mono.Mono, cnt []int) {
	var one int
	var zero int
	source := convert(mono.Just(1).
		DoOnNext(func(v reactor.Any) error {
			one++
			return nil
		}).
		DoOnError(func(e error) {
			zero++
		}))

	source.
		DoOnNext(func(v reactor.Any) error {
			cnt[cntDoOnNext]++
			return nil
		}).
		DoOnError(func(e error) {
			cnt[cntDoOnError]++
			t.Log("[PANIC] DoOnError:", e)
		}).
		DoOnComplete(func() {
			cnt[cntDoOnComplete]++
		}).
		Subscribe(context.Background(),
			reactor.OnNext(func(v reactor.Any) error {
				cnt[cntOnNext]++
				return nil
			}),
			reactor.OnError(func(e error) {
				cnt[cntOnError]++
				t.Log("[PANIC] OnError:", e)
			}),
			reactor.OnComplete(func() {
				cnt[cntOnComplete]++
			}),
		)
	assert.Equal(t, 1, one)
	assert.Equal(t, 0, zero)
}
