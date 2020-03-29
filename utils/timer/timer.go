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
	return &Timer{startTime: time.Now(), lastTime: startTime}
}

func (t *Timer) String() string {
	t.index++
	lastTime := time.Now()
	t.lastTime = lastTime
	if t.index == 1 {
		return fmt.Sprintf("(%v)", time.Since(t.startTime))
	}
	return fmt.Sprintf("%v (#%d %v)", time.Since(lastTime), t.index, time.Since(t.startTime))
}
