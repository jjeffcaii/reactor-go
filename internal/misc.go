package internal

import (
	"errors"
)

var ErrCallOnSubscribeDuplicated = errors.New("call OnSubscribe duplicated")
var EmptySubscription emptySubscription

type emptySubscription struct {
}

func (emptySubscription) Request(n int) {
}

func (emptySubscription) Cancel() {
}

func SafeCloseDone(done chan<- struct{}) (ok bool) {
	defer func() {
		ok = recover() == nil
	}()
	close(done)
	return
}
