package async

import "github.com/michael-reichenauer/gmc/utils/one"

type Async[Result any] struct {
	result      Result
	isResultSet bool
	err         error
	isErrorSet  bool

	errorCallbacks   []func(err error)
	resultCallbacks  []func(r Result)
	finallyCallbacks []func()
	isCompleted      bool

	nextCallbacks []func(r Result, err error)
}

func Run[Result any](runFunc func() (Result, error)) *Async[Result] {
	a := &Async[Result]{}

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

func RunAction(runFunc func() error) *Async[any] {
	a := &Async[any]{}

	go func() {
		err := runFunc()
		one.Do(func() {
			if err != nil {
				a.setError(err)
				return
			}
			a.setResult(nil)
		})
	}()

	return a
}

func RunNoErr[Result any](runFunc func() Result) *Async[Result] {
	a := &Async[Result]{}

	go func() {
		r := runFunc()
		one.Do(func() {
			a.setResult(r)
		})
	}()

	return a
}

func RunActionNoErr(runFunc func()) *Async[any] {
	a := &Async[any]{}

	go func() {
		runFunc()
		one.Do(func() {
			a.setResult(nil)
		})
	}()

	return a
}

func ThenRun[PreResult any, Result any](prev *Async[PreResult], runFunc func(r PreResult) (Result, error)) *Async[Result] {
	a := &Async[Result]{}

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

func (t *Async[Result]) Then(resultCallback func(r Result)) *Async[Result] {
	one.Do(func() {
		// Store callback to be called when/if result is set
		t.resultCallbacks = append(t.resultCallbacks, resultCallback)

		if t.isResultSet {
			// Result already set, call result callback
			t.callResultCallback()
		}
	})

	return t
}

func (t *Async[Result]) Catch(errCallback func(e error)) *Async[Result] {
	one.Do(func() {
		// Store callback to be called when/if error is set
		t.errorCallbacks = append(t.errorCallbacks, errCallback)

		if t.isErrorSet {
			// Error already set, call error callback
			t.callErrCallback()
		}
	})

	return t
}

func (t *Async[Result]) Finally(finallyCallback func()) *Async[Result] {
	one.Do(func() {
		// Store callback to be called when/if result/error is set
		t.finallyCallbacks = append(t.finallyCallbacks, finallyCallback)

		if t.isResultSet || t.isErrorSet {
			// Result or error already set, call finally callback
			t.callFinallyCallback()
		}
	})

	return t
}

func (t *Async[Result]) next(nextCallback func(r Result, err error)) *Async[Result] {
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

func (t *Async[Result]) setResult(result Result) {
	if t.isCompleted {
		// Already completed
		return
	}
	t.result = result
	t.isResultSet = true
	t.isCompleted = true
	t.callResultCallback()
}

func (t *Async[Result]) setError(err error) {
	if t.isCompleted {
		// Already completed
		return
	}

	t.err = err
	t.isErrorSet = true
	t.isCompleted = true
	t.callErrCallback()
}

func (t *Async[Result]) callResultCallback() {
	if t.resultCallbacks != nil {
		callbacks := t.resultCallbacks
		t.resultCallbacks = nil
		for _, callback := range callbacks {
			callback(t.result)
		}
	}

	t.callFinallyCallback()
}

func (t *Async[Result]) callErrCallback() {
	if t.errorCallbacks != nil {
		callbacks := t.errorCallbacks
		t.errorCallbacks = nil
		t.isCompleted = true
		for _, callback := range callbacks {
			callback(t.err)
		}
	}

	t.callFinallyCallback()
}

func (t *Async[Result]) callFinallyCallback() {
	if t.finallyCallbacks != nil {
		callbacks := t.finallyCallbacks
		t.finallyCallbacks = nil
		for _, callback := range callbacks {
			callback()
		}
	}

	t.callNextCallback()
}

func (t *Async[Result]) callNextCallback() {
	if t.nextCallbacks != nil {
		callbacks := t.nextCallbacks
		t.nextCallbacks = nil
		for _, callback := range callbacks {
			callback(t.result, t.err)
		}
	}
}
