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
	p := tuple.NewTuple()
	assert.Equal(t, 0, p.Len(), "len should be zero")

	validate := func(err error) {
		assert.True(t, tuple.IsIndexOutOfBoundsError(err), "should be out of bounds")
	}
	_, err := p.First()
	validate(err)
	_, err = p.Second()
	validate(err)
	_, err = p.Last()
	validate(err)
	_, err = p.Last()
	validate(err)
	_, err = p.Get(0)
	validate(err)
}

func TestTuple_HasError(t *testing.T) {
	tu := tuple.NewTuple(&reactor.Item{V: 1}, &reactor.Item{E: fakeErr})
	assert.True(t, tu.HasError())
	tu = tuple.NewTuple(&reactor.Item{V: 1})
	assert.False(t, tu.HasError())
}

func TestTuple_GetValue(t *testing.T) {
	tu := tuple.NewTuple(&reactor.Item{V: 1}, &reactor.Item{V: 2})
	assert.Equal(t, 1, tu.GetValue(0))
	assert.Equal(t, 2, tu.GetValue(1))
	assert.Nil(t, tu.GetValue(2))
	assert.Nil(t, tu.GetValue(-1))

	tu = tuple.NewTuple(nil)
	assert.Nil(t, tu.GetValue(0))
}

func TestTuple_CollectSlice(t *testing.T) {
	var err error
	tu := tuple.NewTuple(&reactor.Item{V: 1}, &reactor.Item{V: 2})

	err = tu.CollectSlice(nil)
	assert.Error(t, err)

	s := make([]int, 0)

	err = tu.CollectSlice(s)
	assert.Error(t, err)
	err = tu.CollectSlice(&s)
	assert.NoError(t, err)
	assert.Equal(t, []int{1, 2}, s)

	ss := make([]string, 0)
	err = tu.CollectSlice(&ss)
	assert.Error(t, err)

	tu = tuple.NewTuple(nil, &reactor.Item{E: fakeErr}, &reactor.Item{V: 1})
	s = s[:0]
	err = tu.CollectSlice(&s)
	assert.NoError(t, err)
	assert.Equal(t, []int{1}, s)
}
