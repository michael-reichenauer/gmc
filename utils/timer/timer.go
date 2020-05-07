package timer

import (
	"fmt"
	"time"
)

type Timer struct {
	startTime time.Time
	lastTime  time.Time
	index     int
}

func Start() *Timer {
	startTime := time.Now()
	return &Timer{startTime: startTime, lastTime: startTime}
}

func (t *Timer) String() string {
	t.index++
	sinceLast := time.Since(t.lastTime)
	t.lastTime = time.Now()
	if t.index == 1 {
		return fmt.Sprintf("(%v)", time.Since(t.startTime))
	}
	return fmt.Sprintf("(#%d %v) (%v)", t.index, sinceLast, time.Since(t.startTime))
}
