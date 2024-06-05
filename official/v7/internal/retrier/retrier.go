package retrier

import (
	"sync"
	"time"
)

// NewSimpleBackoff creates a SimpleBackoff algorithm with the specified
// list of fixed intervals in milliseconds.
func NewSimpleBackoff(ticks ...int) func(attempt int) time.Duration {
	s := &simpleBackoff{
		ticks: ticks,
	}
	return s.RetryBackoff
}

// simpleBackoff takes a list of fixed values for backoff intervals.
// Each call to RetryBackoff returns the next value from that fixed list.
// After each value is returned, subsequent calls to Next will only return
// the last element.
type simpleBackoff struct {
	sync.Mutex
	ticks []int
}

// RetryBackoff implements a backoff function for SimpleBackoff.
func (b *simpleBackoff) RetryBackoff(retry int) time.Duration {
	b.Lock()
	defer b.Unlock()

	if retry >= len(b.ticks) {
		return 0
	}

	ms := b.ticks[retry]
	return time.Duration(ms) * time.Millisecond
}
