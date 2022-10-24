package one

import (
	"fmt"
	"runtime"
	"time"

	"github.com/michael-reichenauer/gmc/utils"
	"github.com/michael-reichenauer/gmc/utils/log"
)

var inChannel, outChannel = utils.InfiniteChannel()

type funcInfo struct {
	f    func()
	info callerInfo
}

func Run(startFunc func()) {
	Do(startFunc)

	for f := range outChannel {
		if f == nil {
			break
		}
		asFunc(f)()
	}
}

func RunWith(startFunc func(), doWrapper func(f func())) {
	Do(startFunc)
	for f := range outChannel {
		if f == nil {
			break
		}

		doWrapper(asFunc(f))
	}
}

func Do(f func()) {
	if f == nil {
		panic(log.Fatal(fmt.Errorf("nil do function is not supported")))
	}
	ci := caller(2)
	inChannel <- funcInfo{f: f, info: ci}
}

func Close() {
	// close(inChannel) would not handle other Do calls during shutdown
	inChannel <- nil
}

func DoAfter(duration time.Duration, f func()) {
	var t *time.Timer
	t = time.AfterFunc(duration, func() {
		t.Stop()
		Do(f)
	})
}

func asFunc(f interface{}) func() {
	//return f.(func())
	return func() {
		st := time.Now()
		fi := f.(funcInfo)
		fi.f()
		checkDuration(st, fi.info)
	}
}

type callerInfo struct {
	file     string
	line     int
	function string
}

func caller(skip int) callerInfo {
	rpc := make([]uintptr, 1)
	n := runtime.Callers(skip+1, rpc[:])
	if n < 1 {
		return callerInfo{}
	}
	frame, _ := runtime.CallersFrames(rpc).Next()
	return callerInfo{file: frame.File, line: frame.Line, function: frame.Function}
}

func checkDuration(startTime time.Time, info callerInfo) {
	d := time.Since(startTime)
	if d > 10*time.Millisecond {
		log.Warnf("One thread duration (%v) for %s:%d in %s", d, info.file, info.line, info.function)
	}
}
