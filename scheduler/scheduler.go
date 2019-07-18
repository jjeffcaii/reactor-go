package scheduler

type Job func()

type Worker interface {
	Do(Job)
}

type Scheduler interface {
	Worker() Worker
}
