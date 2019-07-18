package scheduler

var immediate Scheduler

func init() {
	immediate = new(immediateScheduler)
}

type immediateScheduler struct {
}

func (s *immediateScheduler) Close() error {
	return nil
}

func (s *immediateScheduler) Do(j Job) {
	j()
}

func (s *immediateScheduler) Worker() Worker {
	return s
}

func Immediate() Scheduler {
	return immediate
}
