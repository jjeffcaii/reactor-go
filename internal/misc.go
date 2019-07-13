package internal

import (
	"errors"
	"fmt"
)

var ErrCallOnSubscribeDuplicated = errors.New("call OnSubscribe duplicated")

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
