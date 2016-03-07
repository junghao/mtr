package mtrapp

import (
	"time"
)

var timers chan Timer

// for aggregating timers
var count = make(map[string]uint64)
var sum = make(map[string]uint64)

func init() {
	timers = make(chan Timer, 300)
}

// TImer is for timing events
type Timer struct {
	start   time.Time
	id      string
	taken   uint64
	stopped bool
}

// Start returns started Timer for Id.
func Start(id string) Timer {
	return Timer{
		start: time.Now().UTC(),
		id:    id,
	}
}

// Stops the timer
func (t *Timer) Stop() {
	t.taken = uint64(time.Since(t.start) / time.Millisecond)
	t.stopped = true
}

// Stops the timer if it is not already stopped.  Tracks the time taken
// in milliseconds.
func (t *Timer) Track() {
	if !t.stopped {
		t.Stop()
	}
	select {
	case timers <- *t:
	default:
	}
}

// Returns the time taken between start and stop in milliseconds.
func (t *Timer) Taken() uint64 {
	return t.taken
}
