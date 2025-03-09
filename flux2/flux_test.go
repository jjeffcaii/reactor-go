package flux

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/jjeffcaii/reactor-go"
)

func TestRange(t *testing.T) {
	Range[int](0, 10).Subscribe(context.TODO(), reactor.OnNext(func(ctx context.Context, next int) error {
		t.Log("Next:", next)
		return nil
	}))
}

func TestFilter(t *testing.T) {
	evens := Range[int](0, 10).Filter(func(ctx context.Context, next int) bool {
		return next&1 == 0
	})

	evens.Subscribe(context.TODO(), reactor.OnNext(func(ctx context.Context, next int) error {
		t.Log("Next:", next)
		return nil
	}))
}

func TestPeek(t *testing.T) {
	Just[int](1, 2, 3, 5, 8, 13).
		DoFinally(func(ctx context.Context, s reactor.SignalType) {
			t.Logf("DoFinally: signal=%v\n", s)
		}).
		DoOnNext(func(ctx context.Context, next int) error {
			t.Log("DoOnNext:", next)
			return nil
		}).
		Subscribe(context.TODO())
}

func TestMap(t *testing.T) {
	origin := Range[int](0, 10)

	origin.
		Map(func(ctx context.Context, next int) (int, error) {
			return next * 10, nil
		}).
		DoOnNext(func(ctx context.Context, next int) error {
			t.Log("DoOnNext:", next)
			return nil
		}).
		Subscribe(context.TODO())

	Map[int, string](
		origin,
		func(ctx context.Context, next int) (string, error) {
			return fmt.Sprintf("Item-%d", next), nil
		}).
		DoOnNext(func(ctx context.Context, next string) error {
			t.Log("DoOnNext:", next)
			return nil
		}).
		Subscribe(context.TODO())
}

func TestCreate(t *testing.T) {
	c := Create[int](func(ctx context.Context, sink Sink[int]) {
		for i := 0; i < 10; i++ {
			sink.Next(ctx, i)
		}
		sink.Complete(ctx)
	})

	c.DoOnNext(func(ctx context.Context, next int) error {
		t.Log("DoOnNext:", next)
		return nil
	}).Subscribe(context.TODO())
}

func TestInterval(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.TODO(), 10*time.Second)
	defer cancel()

	t.Log("Begin:", time.Now())

	Interval2(3*time.Second, time.Second).
		DoOnNext(func(ctx context.Context, next int64) error {
			t.Log("DoOnNext:", time.Now(), next)
			return nil
		}).
		BlockLast(ctx)
}
