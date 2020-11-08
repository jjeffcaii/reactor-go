package reactor

import (
	"errors"
	"fmt"
)

var (
	ErrNegativeRequest    = fmt.Errorf("invalid request: n must be between %d and %d", 1, RequestInfinite)
	ErrSubscribeCancelled = errors.New("subscriber has been cancelled")
)

// IsCancelledError returns true if given error is a cancelled subscribe error.
func IsCancelledError(err error) bool {
	return err == ErrSubscribeCancelled
}
