package flux

import "io"

type queue interface {
	io.Closer
	offer(interface{})
	poll() (interface{}, bool)
	size() int
}

type simpleQueue struct {
	c chan interface{}
}

func (q simpleQueue) size() int {
	return len(q.c)
}

func (q simpleQueue) Close() (err error) {
	close(q.c)
	return
}

func (q simpleQueue) offer(v interface{}) {
	q.c <- v
}

func (q simpleQueue) poll() (v interface{}, ok bool) {
	select {
	case v, ok = <-q.c:
		return
	default:
		return
	}
}

func newQueue(size int) queue {
	return simpleQueue{
		c: make(chan interface{}, size),
	}
}
