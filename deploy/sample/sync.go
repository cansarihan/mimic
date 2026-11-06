package sample

import "sync"

type counter struct {
	mu sync.Mutex
	n  int
}

// bump locks but never unlocks — a deadlock on the next caller.
func (c *counter) bump() {
	c.mu.Lock()
	c.n++
}

// collect discards the appended slice header.
func collect(seen []string, item string) {
	_ = append(seen, item)
}
