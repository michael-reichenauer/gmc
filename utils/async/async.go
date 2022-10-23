package async

import (
	"fmt"

	"github.com/michael-reichenauer/gmc/utils/log"
	"github.com/michael-reichenauer/gmc/utils/one"
)

type Task[Result any] interface {
	Then(resultCallback func(r Result)) Task[Result]
	Catch(errCallback func(e error)) Task[Result]
	Finally(finallyCallback func()) Task[Result]
	next(nextCallback func(r Result, err error)) Task[Result]
}

type task[Result any] struct {
	result          Result
	err             error
	isCompleted     bool
	errorCallback   func(err error)
	resultCallback  func(r Result)
	finallyCallback func()

	nextCallbacks []func(r Result, err error)
}

func Run[Result any](runFunc func() (Result, error)) Task[Result] {
	a := &task[Result]{}

	go func() {
		r, err := runFunc()
		one.Do(func() {
			if err != nil {
				a.setError(err)
				return
			}

			a.setResult(r)
		})
	}()

	return a
}

func RunNoErr[Result any](runFunc func() Result) Task[Result] {
	return Run(func() (Result, error) { return runFunc(), nil })
}

func RunAction(runFunc func() error) Task[any] {
	return Run(func() (any, error) { return nil, runFunc() })
}

func RunActionNoErr(runFunc func()) Task[any] {
	return Run(func() (any, error) {
		runFunc()
		return nil, nil
	})
}

func ThenRun[PreResult any, Result any](prev Task[PreResult], runFunc func(r PreResult) (Result, error)) Task[Result] {
	a := &task[Result]{}

	prev.next(func(preResult PreResult, preErr error) {
		if preErr != nil {
			a.setError(preErr)
			return
		}

		go func() {
			r, err := runFunc(preResult)
			one.Do(func() {
				if err != nil {
					a.setError(err)
					return
				}

				a.setResult(r)
			})
		}()
	})

	return a
}

func ThenRunNoErr[PreResult any, Result any](prev Task[PreResult], runFunc func(r PreResult) Result) Task[any] {
	return ThenRun(prev, func(r PreResult) (any, error) { return runFunc(r), nil })
}

func ThenRunAction[PreResult any](prev Task[PreResult], runFunc func(r PreResult) error) Task[any] {
	return ThenRun(prev, func(r PreResult) (any, error) { return nil, runFunc(r) })
}

func ThenRunActionNoErr[PreResult any](prev Task[PreResult], runFunc func(r PreResult)) Task[any] {
	return ThenRun(prev, func(pr PreResult) (any, error) {
		runFunc(pr)
		return nil, nil
	})
}

func (t *task[Result]) Then(callback func(r Result)) Task[Result] {
	one.Do(func() {
		if t.resultCallback != nil {
			panic(log.Fatal(fmt.Errorf("only one Then() call allows on a task")))
		}
		// Store callback to be called when/if result is set
		t.resultCallback = callback

		if t.isCompleted {
			// Result already set, call result callback
			t.callResultCallback()
		}
	})

	return t
}

func (t *task[Result]) Catch(callback func(e error)) Task[Result] {
	one.Do(func() {
		if t.errorCallback != nil {
			panic(log.Fatal(fmt.Errorf("only one Catch() call allows on a task")))
		}
		// Store callback to be called when/if error is set
		t.errorCallback = callback

		if t.isCompleted {
			// Error already set, call error callback
			t.callErrCallback()
		}
	})

	return t
}

func (t *task[Result]) Finally(callback func()) Task[Result] {
	one.Do(func() {
		if t.finallyCallback != nil {
			panic(log.Fatal(fmt.Errorf("only one Finally() call allows on a task")))
		}
		// Store callback to be called when/if result/error is set
		t.finallyCallback = callback

		if t.isCompleted {
			// Result or error already set, call finally callback
			t.callFinallyCallback()
		}
	})

	return t
}

func (t *task[Result]) next(nextCallback func(r Result, err error)) Task[Result] {
	one.Do(func() {
		// Store callback to be called when/if result is set
		t.nextCallbacks = append(t.nextCallbacks, nextCallback)

		if t.isCompleted {
			// Result/error already set, call result callback
			t.callNextCallback()
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
	t.callResultCallback()
}

func (t *task[Result]) setError(err error) {
	if t.isCompleted {
		return
	}

	t.err = err
	t.isCompleted = true
	t.callErrCallback()
}

func (t *task[Result]) callResultCallback() {
	if t.resultCallback != nil {
		callback := t.resultCallback
		t.resultCallback = nil
		callback(t.result)
	}

	t.callFinallyCallback()
}

func (t *task[Result]) callErrCallback() {
	if t.errorCallback != nil {
		callback := t.errorCallback
		t.errorCallback = nil
		t.isCompleted = true
		callback(t.err)
	}

	t.callFinallyCallback()
}

func (t *task[Result]) callFinallyCallback() {
	if t.finallyCallback != nil {
		callback := t.finallyCallback
		t.finallyCallback = nil
		callback()
	}

	t.callNextCallback()
}

func (t *task[Result]) callNextCallback() {
	if t.nextCallbacks != nil {
		callbacks := t.nextCallbacks
		t.nextCallbacks = nil
		for _, callback := range callbacks {
			callback(t.result, t.err)
		}
	}
}
