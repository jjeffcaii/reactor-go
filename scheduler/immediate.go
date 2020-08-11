package scheduler

const _immediateName = "immediate"

var _immediate Scheduler

func init() {
	_immediate = immediateScheduler{}
}

type immediateScheduler struct {
}

func (i immediateScheduler) Name() string {
	return _immediateName
}

func (i immediateScheduler) Close() error {
	return nil
}

func (i immediateScheduler) Do(job Task) error {
	job()
	return nil
}

func (i immediateScheduler) Worker() Worker {
	return i
}

// Immediate returns a scheduler which schedule tasks _immediate.
func Immediate() Scheduler {
	return _immediate
}
