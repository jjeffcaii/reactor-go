package mono_test

import (
	"context"
	"testing"

	"github.com/jjeffcaii/reactor-go/mono"
)

func BenchmarkJust(b *testing.B) {
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			mono.Just(1).Subscribe(context.Background())
		}
	})
}

func BenchmarkCreate(b *testing.B) {
	gen := func(i context.Context, sink mono.Sink) {
		sink.Success(1)
	}
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			mono.Create(gen).Subscribe(context.Background())
		}
	})
}

func BenchmarkJustOneshot(b *testing.B) {
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			mono.JustOneshot(1).Subscribe(context.Background())
		}
	})
}

func BenchmarkCreateOneshot(b *testing.B) {
	gen := func(i context.Context, sink mono.Sink) {
		sink.Success(1)
	}
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			mono.CreateOneshot(gen).Subscribe(context.Background())
		}
	})
}
