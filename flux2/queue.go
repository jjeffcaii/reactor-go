package flux

type queue[T any] struct {
	c chan T
}

func (q queue[T]) size() int {
	return len(q.c)
}

func (q queue[T]) Close() (err error) {
	close(q.c)
	return
}

func (q queue[T]) offer(v T) {
	q.c <- v
}

func (q queue[T]) poll() (v T, ok bool) {
	select {
	case v, ok = <-q.c:
		return
	default:
		return
	}
}

func newQueue[T any](size int) queue[T] {
	return queue[T]{
		c: make(chan T, size),
	}
}
