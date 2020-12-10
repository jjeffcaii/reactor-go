package tuple

import (
	"errors"

	"github.com/jjeffcaii/reactor-go"
)

var empty Tuple = (tuple)(nil)

var errIndexOutOfBounds = errors.New("index out of bounds")

// Tuple is a container of multi elements.
type Tuple interface {
	// First returns the first element in the Tuple.
	First() (reactor.Any, error)
	// Second returns the second element in the Tuple.
	Second() (reactor.Any, error)
	// Last returns the last element in the Tuple.
	Last() (reactor.Any, error)
	// Get returns the element in the Tuple with given index.
	Get(index int) (reactor.Any, error)
	// GetValue returns value of the element in the Tuple with given index.
	GetValue(index int) reactor.Any
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
	return tuple(items)
}

type tuple []*reactor.Item

func (t tuple) GetValue(index int) reactor.Any {
	value, _ := t.Get(index)
	return value
}

func (t tuple) checkItem(item *reactor.Item) (reactor.Any, error) {
	if item != nil {
		return item.V, item.E
	}
	return nil, nil
}

func (t tuple) First() (reactor.Any, error) {
	if len(t) < 1 {
		return nil, errIndexOutOfBounds
	}
	return t.checkItem(t[0])
}

func (t tuple) Second() (reactor.Any, error) {
	if len(t) < 2 {
		return nil, errIndexOutOfBounds
	}
	return t.checkItem(t[1])
}

func (t tuple) Last() (reactor.Any, error) {
	if len(t) < 1 {
		return nil, errIndexOutOfBounds
	}
	return t.checkItem(t[len(t)-1])
}

func (t tuple) ForEachWithIndex(callback func(v reactor.Any, e error, index int) (ok bool)) {
	for i := 0; i < len(t); i++ {
		value, err := t.checkItem(t[i])
		if !callback(value, err, i) {
			break
		}
	}
}

func (t tuple) ForEach(f func(v reactor.Any, e error) bool) {
	for _, next := range t {
		value, err := t.checkItem(next)
		if !f(value, err) {
			break
		}
	}
}

func (t tuple) Len() int {
	return len(t)
}

func (t tuple) Get(index int) (reactor.Any, error) {
	if index < 0 || index >= len(t) {
		return nil, errIndexOutOfBounds
	}
	return t.checkItem(t[index])
}
