package bttt

import (
	"time"
)

type _Timer struct {
	start    time.Time
	duration time.Duration
}

func _NewTimer() *_Timer {
	return &_Timer{time.Now(), 1000}
}

// Check if this timer has ended
func (t *_Timer) IsEnd() bool {
	return time.Since(t.start) >= t.duration
}

// Set the 'start' as now
func (t *_Timer) Reset() {
	t.start = time.Now()
}

// Get the start time
func (t *_Timer) Start() time.Time {
	return t.start
}

// In milliseconds
func (t *_Timer) Movetime(movetime int) {
	t.duration = time.Duration(movetime) * time.Millisecond
}
