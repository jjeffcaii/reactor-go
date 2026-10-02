package flux

type queue struct {
	c chan Any
	// done signals termination; c is never closed.
	done chan struct{}
}

func (q queue) size() int {
	return len(q.c)
}

func (q queue) Close() (err error) {
	close(q.done)
	return
}

// offer enqueues v and reports whether it was accepted. A full buffer blocks
// until space frees up or termination wins, so it never panics.
func (q queue) offer(v Any) bool {
	select {
	case <-q.done:
		return false
	default:
	}
	select {
	case q.c <- v:
		return true
	case <-q.done:
		return false
	}
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
		c:    make(chan Any, size),
		done: make(chan struct{}),
	}
}
