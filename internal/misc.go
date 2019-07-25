package internal

import (
	"errors"
	"fmt"

	rs "github.com/jjeffcaii/reactor-go"
)

var (
	ErrCallOnSubscribeDuplicated                 = errors.New("call OnSubscribe duplicated")
	EmptySubscription            rs.Subscription = &emptySubscription{}
)

func TryRecoverError(re interface{}) error {
	if re == nil {
		return nil
	}
	switch e := re.(type) {
	case error:
		return e
	case string:
		return errors.New(e)
	default:
		return fmt.Errorf("%s", e)
	}
}

type emptySubscription struct {
}

func (emptySubscription) Request(n int) {
}

func (emptySubscription) Cancel() {
}
