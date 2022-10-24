package async

import (
	"fmt"

	"github.com/michael-reichenauer/gmc/utils/log"
	"github.com/michael-reichenauer/gmc/utils/one"
)

type Task[Result any] interface {
	Then(resultCallback func(r Result)) Task[Result]
	Catch(errCallback func(err error)) Task[Result]
	Finally(finallyCallback func()) Task[Result]
	onCompleted(nextCallback func(r Result, err error)) Task[Result]
}

type task[Result any] struct {
	result          Result
	err             error
	isCompleted     bool
	thenCallback    func(r Result)
	catchCallback   func(err error)
	finallyCallback func()

	completedCallbacks []func(r Result, err error)
}

func RunRE[Result any](taskFunc func() (Result, error)) Task[Result] {
	task := &task[Result]{}
	task.run(taskFunc)
	return task
}

func ThenRunRE[PrevResult any, Result any](prevTask Task[PrevResult], taskFunc func(r PrevResult) (Result, error)) Task[Result] {
	task := &task[Result]{}

	// Schedule this task to run once previous task has completed
	prevTask.onCompleted(func(prevResult PrevResult, prevErr error) {
		if prevErr != nil {
			// if previous task fails, then this task fails as well
			task.setError(prevErr)
			return
		}

		// No previous error, so this task can run
		task.run(func() (Result, error) { return taskFunc(prevResult) })
	})

	return task
}

func RunR[Result any](runFunc func() Result) Task[Result] {
	return RunRE(func() (Result, error) { return runFunc(), nil })
}

func RunE(runFunc func() error) Task[any] {
	return RunRE(func() (any, error) { return nil, runFunc() })
}

func Run(runFunc func()) Task[any] {
	return RunRE(func() (any, error) {
		runFunc()
		return nil, nil
	})
}

func ThenRunR[PreResult any, Result any](prev Task[PreResult], runFunc func(r PreResult) Result) Task[any] {
	return ThenRunRE(prev, func(r PreResult) (any, error) { return runFunc(r), nil })
}

func ThenRunE[PreResult any](prev Task[PreResult], runFunc func(r PreResult) error) Task[any] {
	return ThenRunRE(prev, func(r PreResult) (any, error) { return nil, runFunc(r) })
}

func ThenRun[PreResult any](prev Task[PreResult], runFunc func(r PreResult)) Task[any] {
	return ThenRunRE(prev, func(pr PreResult) (any, error) {
		runFunc(pr)
		return nil, nil
	})
}

func (t *task[Result]) Then(callback func(r Result)) Task[Result] {
	one.Do(func() {
		if t.thenCallback != nil {
			panic(log.Fatal(fmt.Errorf("only one Then() call allows on a task")))
		}
		// Store callback to be called when result is set
		t.thenCallback = callback

		if t.isCompleted {
			// Result already set, call result callback
			t.callThenCallback()
		}
	})

	return t
}

func (t *task[Result]) Catch(callback func(err error)) Task[Result] {
	one.Do(func() {
		if t.catchCallback != nil {
			panic(log.Fatal(fmt.Errorf("only one Catch() call allows on a task")))
		}
		// Store callback to be called when error is set
		t.catchCallback = callback

		if t.isCompleted {
			// Error already set, call error callback
			t.callCatchCallback()
		}
	})

	return t
}

func (t *task[Result]) Finally(callback func()) Task[Result] {
	one.Do(func() {
		if t.finallyCallback != nil {
			panic(log.Fatal(fmt.Errorf("only one Finally() call allows on a task")))
		}
		// Store callback to be called when result/error is set
		t.finallyCallback = callback

		if t.isCompleted {
			// Result or error already set, call finally callback
			t.callFinallyCallback()
		}
	})

	return t
}

func (t *task[Result]) run(runFunc func() (Result, error)) {
	go func() {
		r, err := runFunc()
		one.Do(func() {
			if err != nil {
				t.setError(err)
				return
			}

			t.setResult(r)
		})
	}()
}

func (t *task[Result]) onCompleted(completedCallback func(r Result, err error)) Task[Result] {
	one.Do(func() {
		// Store callback to next task to be called when task is completed
		t.completedCallbacks = append(t.completedCallbacks, completedCallback)

		if t.isCompleted {
			// Result/error already set, call completed callback
			t.callCompletedCallbacks()
		}
	})

	return t
}

func (t *task[Result]) setResult(result Result) {
	if t.isCompleted {
		return
	}

	t.result = result
	t.isCompleted = true

	t.callThenCallback()
}

func (t *task[Result]) setError(err error) {
	if t.isCompleted {
		return
	}

	t.err = err
	t.isCompleted = true

	t.callCatchCallback()
}

func (t *task[Result]) callThenCallback() {
	if t.thenCallback != nil {
		t.thenCallback(t.result)
	}

	t.callFinallyCallback()
}

func (t *task[Result]) callCatchCallback() {
	if t.catchCallback != nil {
		t.catchCallback(t.err)
	}

	t.callFinallyCallback()
}

func (t *task[Result]) callFinallyCallback() {
	if t.finallyCallback != nil {
		t.finallyCallback()
	}

	t.callCompletedCallbacks()
}

func (t *task[Result]) callCompletedCallbacks() {
	if t.completedCallbacks != nil {
		callbacks := t.completedCallbacks
		t.completedCallbacks = nil
		for _, callback := range callbacks {
			callback(t.result, t.err)
		}
	}
}
