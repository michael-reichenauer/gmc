package timer

import (
	"fmt"
	"time"
)

type Timer struct {
	time time.Time
}

func Start() *Timer {
	return &Timer{time: time.Now()}
}

func (t *Timer) String() string {
	return fmt.Sprintf("%v", time.Since(t.time))
}
