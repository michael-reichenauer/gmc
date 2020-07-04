package utils

type Actor struct {
	in   chan<- interface{}
	out  <-chan interface{}
	done chan struct{}
}

func newActor() *Actor {
	t := &Actor{done: make(chan struct{})}
	t.in, t.out = InfiniteChannel()

	go t.workRoutine()
	return t
}

// Do runs the specified func in the using the actor routine
func (t *Actor) Do(f func()) {
	select {
	case t.in <- f:
	case <-t.done:
	}
}

// Do runs the specified func in the using the actor routine and returns the result
func (t *Actor) DoFunc(f func() interface{}) interface{} {
	result := make(chan interface{})
	select {
	case t.in <- func() { result <- f() }:
	case <-t.done:
		return nil
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
	close(t.in)
}

func (t *Actor) workRoutine() {
	for {
		select {
		case w, ok := <-t.out:
			if !ok {
				return
			}
			work := w.(func())
			work()
		case <-t.done:
			return
		}
	}
}
