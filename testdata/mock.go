package testdata

import (
	"context"

	"github.com/jjeffcaii/reactor-go"
)

func MockRequestInfinite(ctx context.Context, su reactor.Subscription) {
	su.Request(ctx, reactor.RequestInfinite)
}
