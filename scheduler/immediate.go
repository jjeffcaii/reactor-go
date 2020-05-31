package scheduler

var immediate Scheduler

func init() {
	immediate = new(immediateScheduler)
}

type immediateScheduler struct {
}

func (i *immediateScheduler) Close() error {
	return nil
}

func (i *immediateScheduler) Do(job Job) {
	job()
}

func (i *immediateScheduler) Worker() Worker {
	return i
}

func Immediate() Scheduler {
	return immediate
}
