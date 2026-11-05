package engine

import (
	"context"
	"testing"
)

func TestCancelsStopsQueuedAndRunningJobs(t *testing.T) {
	c := NewCancels()
	if !c.begin(`C:\clips\a.mp4`, func() {}) {
		t.Fatal("begin rejected a fresh file")
	}
	c.end(`C:\clips\a.mp4`)

	c.Cancel(`C:\clips\a.mp4`)
	if c.begin(`C:\clips\a.mp4`, func() {}) {
		t.Fatal("begin accepted a file cancelled before it started")
	}

	runningCtx, runningCancel := context.WithCancel(context.Background())
	defer runningCancel()
	if !c.begin(`C:\clips\b.mp4`, runningCancel) {
		t.Fatal("begin rejected a fresh file")
	}
	c.Cancel(`C:\clips\b.mp4`)
	if runningCtx.Err() == nil {
		t.Fatal("cancelling a running file did not stop its context")
	}
	c.end(`C:\clips\b.mp4`)
}
