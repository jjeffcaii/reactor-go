package stopwatch

import "time"

type Stopwatch struct {
	prev time.Time
}

func NewStopwatch() *Stopwatch {
	return &Stopwatch{
		prev: time.Now(),
	}
}

func (sw Stopwatch) Elapsed() time.Duration {
	return time.Since(sw.prev)
}

func (sw *Stopwatch) Reset() {
	sw.prev = time.Now()
}
