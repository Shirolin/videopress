package engine

import (
	"context"
	"path/filepath"
	"sync"
)

// Cancels tracks per-file cancellation of a running batch. The zero value is
// not usable; call NewCancels.
type Cancels struct {
	mu        sync.Mutex
	cancelled map[string]bool
	running   map[string]context.CancelFunc
}

// NewCancels creates a cancellation registry for one batch.
func NewCancels() *Cancels {
	return &Cancels{
		cancelled: map[string]bool{},
		running:   map[string]context.CancelFunc{},
	}
}

// Cancel stops the job for path, or keeps a queued job from starting.
func (c *Cancels) Cancel(path string) {
	key := filepath.Clean(path)
	c.mu.Lock()
	c.cancelled[key] = true
	cancel := c.running[key]
	c.mu.Unlock()
	if cancel != nil {
		cancel()
	}
}

// begin registers a running job; false means it was cancelled before start.
func (c *Cancels) begin(path string, cancel context.CancelFunc) bool {
	key := filepath.Clean(path)
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.cancelled[key] {
		return false
	}
	c.running[key] = cancel
	return true
}

// end drops the running entry once its job finished.
func (c *Cancels) end(path string) {
	key := filepath.Clean(path)
	c.mu.Lock()
	delete(c.running, key)
	c.mu.Unlock()
}
