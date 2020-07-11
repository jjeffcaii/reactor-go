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
	// Worker returns next worker.
	Worker() Worker
}
