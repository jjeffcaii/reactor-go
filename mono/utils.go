package mono

import (
	"context"

	rs "github.com/jjeffcaii/reactor-go"
)

func toBlock(ctx context.Context, pub rs.Publisher) (v interface{}, err error) {
	chValue := make(chan interface{}, 1)
	chError := make(chan error, 1)
	pub.Subscribe(ctx,
		rs.OnNext(func(s rs.Subscription, v interface{}) {
			chValue <- v
		}),
		rs.OnComplete(func() {
			close(chValue)
		}),
		rs.OnError(func(e error) {
			chError <- e
			close(chError)
		}),
	)
	select {
	case v = <-chValue:
		close(chError)
	case err = <-chError:
		close(chValue)
	}
	return
}
