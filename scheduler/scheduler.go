package scheduler

import "io"

// Task defines task function.
type Task func()

// Worker is used to execute a task.
type Worker interface {
	// Do executes a task.
	Do(Task)
}

// Scheduler schedule tasks.
type Scheduler interface {
	io.Closer
	// Name return the name of scheduler.
	Name() string
	// Worker returns next worker.
	Worker() Worker
}
