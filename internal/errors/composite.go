package errors

import (
	"strings"
)

type compositeError struct {
	errors []error
	idx    []int
}

func (c compositeError) Len() int {
	return len(c.idx)
}

func (c compositeError) At(i int) error {
	if i < 0 || i >= len(c.idx) {
		return nil
	}
	return c.errors[i]
}

func (c compositeError) Error() string {
	if len(c.errors) == 1 {
		return c.errors[0].Error()
	}
	b := strings.Builder{}
	b.WriteString(c.errors[0].Error())
	for i := 1; i < len(c.errors); i++ {
		b.WriteByte('\n')
		b.WriteString(c.errors[i].Error())
	}
	return b.String()
}

func NewCompositeError(errors []error, indexes []int) error {
	return compositeError{
		errors: errors,
		idx:    indexes,
	}
}
