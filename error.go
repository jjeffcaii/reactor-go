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
	if err == ErrSubscribeCancelled {
		return true
	}
	if _, ok := err.(contextError); ok {
		return true
	}
	if _, ok := err.(*contextError); ok {
		return true
	}
	return false
}

type CompositeError interface {
	error
	Len() int
	At(int) error
}

type contextError struct {
	e error
}

func NewContextError(err error) error {
	return contextError{e: err}
}

func (c contextError) Error() string {
	return c.e.Error()
}

func (c contextError) Cause() error {
	return c.e
}
