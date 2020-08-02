package flux

type queue struct {
	c chan Any
}

func (q queue) size() int {
	return len(q.c)
}

func (q queue) Close() (err error) {
	close(q.c)
	return
}

func (q queue) offer(v Any) {
	q.c <- v
}

func (q queue) poll() (v Any, ok bool) {
	select {
	case v, ok = <-q.c:
		return
	default:
		return
	}
}

func newQueue(size int) queue {
	return queue{
		c: make(chan Any, size),
	}
}
