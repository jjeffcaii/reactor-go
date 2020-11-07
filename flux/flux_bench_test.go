package flux_test

import (
	"context"
	"testing"

	"github.com/jjeffcaii/reactor-go/flux"
)

func BenchmarkBlockLast(b *testing.B) {
	f := flux.Just(1, 2, 3, 4)
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, _ = f.BlockLast(context.Background())
		}
	})
}
