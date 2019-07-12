package scheduler

import "io"

type Job func()

type Worker interface {
	io.Closer
	Do(Job)
}

type Scheduler interface {
	Worker() Worker
}
