package scheduler

const _parallelName = "parallel"

var _parallel Scheduler

func init() {
	_parallel = parallelScheduler{}
}

type parallelScheduler struct {
}

func (p parallelScheduler) Name() string {
	return _parallelName
}

func (p parallelScheduler) Close() error {
	return nil
}

func (p parallelScheduler) Do(j Task) {
	go j()
}

func (p parallelScheduler) Worker() Worker {
	return p
}

// Parallel returns scheduler which schedule tasks in _parallel using native goroutines.
func Parallel() Scheduler {
	return _parallel
}
