package flux

import "errors"

const (
	statCancel   = -1
	statError    = -2
	statComplete = 2
)

var errSubscribeOnce = errors.New("only one subscriber is allow")
