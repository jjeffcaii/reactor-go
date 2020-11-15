package tuple_test

import (
	"errors"
	"testing"

	"github.com/jjeffcaii/reactor-go"
	"github.com/jjeffcaii/reactor-go/tuple"
	"github.com/stretchr/testify/assert"
)

var fakeErr = errors.New("fake error")

func TestNewTuple(t *testing.T) {
	items := []*reactor.Item{{V: 1}, {V: 2}, {E: fakeErr}}
	tup := tuple.NewTuple(items...)
	assert.Equal(t, 3, tup.Len())
	v, err := tup.First()
	assert.NoError(t, err)
	assert.Equal(t, items[0].V, v)
	v, _ = tup.Second()
	assert.Equal(t, items[1].V, v)
	_, err = tup.Last()
	assert.Error(t, err)
	assert.Equal(t, fakeErr, err)

	v, _ = tup.Get(0)
	assert.Equal(t, items[0].V, v)
	_, err = tup.Get(-1)
	assert.True(t, tuple.IsIndexOutOfBoundsError(err))
	_, err = tup.Get(tup.Len())
	assert.True(t, tuple.IsIndexOutOfBoundsError(err))

	var visits int
	tup.ForEach(func(v reactor.Any, e error) bool {
		visits++
		return true
	})
	assert.Equal(t, tup.Len(), visits)

	visits = 0
	tup.ForEach(func(v reactor.Any, e error) (ok bool) {
		visits++
		return
	})
	assert.Equal(t, 1, visits)

	visits = 0
	tup.ForEachWithIndex(func(v reactor.Any, e error, index int) (ok bool) {
		ok = true
		visits++
		return
	})
	assert.Equal(t, tup.Len(), visits)

	visits = 0
	tup.ForEachWithIndex(func(v reactor.Any, e error, index int) (ok bool) {
		visits++
		return
	})
	assert.Equal(t, 1, visits)
}

func TestEmptyTuple(t *testing.T) {
	tuple.NewTuple()
}
