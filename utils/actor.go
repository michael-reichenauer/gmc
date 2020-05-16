package utils

type Actor struct {
	work chan func()
	done chan struct{}
}

func newActor() *Actor {
	t := &Actor{work: make(chan func()), done: make(chan struct{})}
	go t.workRoutine()
	return t
}

// Do runs the specified func in the using the actor routine
func (t *Actor) Do(f func()) {
	select {
	case t.work <- f:
	case <-t.done:
	}
}

// Do runs the specified func in the using the actor routine and returns the result
func (t *Actor) DoFunc(f func() interface{}) interface{} {
	result := make(chan interface{})
	select {
	case t.work <- func() { result <- f() }:
	case <-t.done:
	}
	return <-result
}

// WaitAllDone waits until all scheduled work has been done
func (t *Actor) WaitAllDone() {
	done := make(chan struct{})
	t.Do(func() { close(done) })
	<-done
}

func (t *Actor) Close() {
	close(t.done)
}

func (t *Actor) workRoutine() {
	for {
		select {
		case work := <-t.work:
			work()
		case <-t.done:
			return
		}
	}
}
