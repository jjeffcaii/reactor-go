package internal

import (
	"context"
	"errors"
)

var ErrCallOnSubscribeDuplicated = errors.New("call OnSubscribe duplicated")

type EmptySubscription struct {
}

func (EmptySubscription) Request(ctx context.Context, n int) {
}

func (EmptySubscription) Cancel(ctx context.Context) {
}

func SafeCloseDone(done chan<- struct{}) (ok bool) {
	defer func() {
		ok = recover() == nil
	}()
	close(done)
	return
}
