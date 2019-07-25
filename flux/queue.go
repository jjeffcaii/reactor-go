package flux

type queue struct {
	c chan interface{}
}

func (q queue) size() int {
	return len(q.c)
}

func (q queue) Close() (err error) {
	close(q.c)
	return
}

func (q queue) offer(v interface{}) {
	q.c <- v
}

func (q queue) poll() (v interface{}, ok bool) {
	select {
	case v, ok = <-q.c:
		return
	default:
		return
	}
}

func newQueue(size int) queue {
	return queue{
		c: make(chan interface{}, size),
	}
}
