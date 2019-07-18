package scheduler

import "io"

type Job func()

type Worker interface {
	Do(Job)
}

type Scheduler interface {
	io.Closer
	Worker() Worker
}
