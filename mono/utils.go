package mono

import (
	"context"
	"errors"
)

var empty Mono = wrapper{newMonoJust(nil)}
var errJustNilValue = errors.New("require non nil value")

func Empty() Mono {
	return empty
}

func JustOrEmpty(v interface{}) Mono {
	if v == nil {
		return empty
	}
	return Just(v)
}

func Just(v interface{}) Mono {
	if v == nil {
		panic(errJustNilValue)
	}
	return wrapper{newMonoJust(v)}
}

func Create(gen func(context.Context, Sink)) Mono {
	return wrapper{newMonoCreate(gen)}
}
