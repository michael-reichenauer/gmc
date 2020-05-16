package utils

import (
	"github.com/stretchr/testify/assert"
	"sync/atomic"
	"testing"
)

func TestActor_Do(t *testing.T) {
	a := newActor()
	defer a.Close()

	var sum int32
	a.Do(func() { atomic.AddInt32(&sum, 1) })
	a.Do(func() { atomic.AddInt32(&sum, 2) })
	a.Do(func() { atomic.AddInt32(&sum, 3) })
	a.Do(func() { atomic.AddInt32(&sum, 4) })

	a.WaitAllDone()
	assert.Equal(t, 10, int(atomic.LoadInt32(&sum)))
}

func TestActor_DoFunc(t *testing.T) {
	a := newActor()
	defer a.Close()

	assert.Equal(t, 10, a.DoFunc(func() interface{} { return 10 }).(int))
}
