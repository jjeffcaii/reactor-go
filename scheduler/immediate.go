package scheduler

var (
	immediate = immediateScheduler{}
)

type immediateWorker struct {
}

func (immediateWorker) Close() error {
	return nil
}

func (immediateWorker) Do(job Job) {
	job()
}

type immediateScheduler struct {
	w immediateWorker
}

func (s immediateScheduler) Worker() Worker {
	return s.w
}

func Immediate() Scheduler {
	return immediate
}
