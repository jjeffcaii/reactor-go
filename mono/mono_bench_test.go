package mono_test

import (
	"context"
	"testing"

	"github.com/jjeffcaii/reactor-go/mono"
)

func BenchmarkZip(b *testing.B) {
	first := mono.Just(1)
	second := mono.Just(2)
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			mono.Zip(first, second).Subscribe(context.Background())
		}
	})
}

func BenchmarkJust(b *testing.B) {
	j := mono.Just(1)
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			j.Subscribe(context.Background())
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

func BenchmarkProcessor(b *testing.B) {
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			p, s, d := mono.NewProcessor(nil, nil)
			s.Success(1)
			p.Subscribe(context.Background())
			d.Dispose()
		}
	})
}
