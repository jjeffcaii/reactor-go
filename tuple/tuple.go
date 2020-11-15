package tuple

import (
	"errors"

	"github.com/jjeffcaii/reactor-go"
)

var empty Tuple = &tuple{}

var errIndexOutOfBounds = errors.New("index out of bounds")

var _ Tuple = (*tuple)(nil)

// Tuple is a container of multi elements.
type Tuple interface {
	// First returns the first element in the Tuple.
	First() (reactor.Any, error)
	// Second returns the second element in the Tuple.
	Second() (reactor.Any, error)
	// Last returns the last element in the Tuple.
	Last() (reactor.Any, error)
	// Get returns the N element in the Tuple with given index.
	Get(index int) (reactor.Any, error)
	// Len returns the length of Tuple.
	Len() int
	// ForEach execute callback for each element in the Tuple.
	// If ok returns false, current loop will be broken.
	ForEach(callback func(v reactor.Any, e error) (ok bool))
	// ForEachWithIndex execute callback for each element and index in the Tuple.
	// If ok returns false, current loop will be broken.
	ForEachWithIndex(callback func(v reactor.Any, e error, index int) (ok bool))
}

// IsIndexOutOfBoundsError returns true if input error is type of "IndexOutOfBounds".
func IsIndexOutOfBoundsError(err error) bool {
	return err == errIndexOutOfBounds
}

// NewTuple returns a new Tuple.
func NewTuple(items ...*reactor.Item) Tuple {
	if len(items) < 1 {
		return empty
	}
	return tuple{inner: items}
}

type tuple struct {
	inner []*reactor.Item
}

func (t tuple) First() (reactor.Any, error) {
	if len(t.inner) < 1 {
		return nil, errIndexOutOfBounds
	}
	next := t.inner[0]
	return next.V, next.E
}

func (t tuple) Second() (reactor.Any, error) {
	if len(t.inner) < 2 {
		return nil, errIndexOutOfBounds
	}
	next := t.inner[1]
	return next.V, next.E
}

func (t tuple) Last() (reactor.Any, error) {
	if len(t.inner) < 1 {
		return nil, errIndexOutOfBounds
	}
	next := t.inner[len(t.inner)-1]
	return next.V, next.E
}

func (t tuple) ForEachWithIndex(callback func(v reactor.Any, e error, index int) (ok bool)) {
	for i := 0; i < len(t.inner); i++ {
		if !callback(t.inner[i].V, t.inner[i].E, i) {
			break
		}
	}
}

func (t tuple) ForEach(f func(any reactor.Any, err error) bool) {
	for _, next := range t.inner {
		if !f(next.V, next.E) {
			break
		}
	}
}

func (t tuple) Len() int {
	return len(t.inner)
}

func (t tuple) Get(index int) (reactor.Any, error) {
	if index < 0 || index >= len(t.inner) {
		return nil, errIndexOutOfBounds
	}
	item := t.inner[index]
	return item.V, item.E
}
