package reactor

import (
	"errors"
	"fmt"
)

var (
	ErrNegativeRequest    = fmt.Errorf("invalid request: n must be between %d and %d", 1, RequestInfinite)
	ErrSubscribeCancelled = errors.New("subscriber has been cancelled")
)

func IsCancelledError(err error) bool {
	return err == ErrSubscribeCancelled
}

var CoreCounts int32
