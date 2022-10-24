package one

import (
	"fmt"
	"time"

	"github.com/michael-reichenauer/gmc/utils"
	"github.com/michael-reichenauer/gmc/utils/log"
)

var inChannel, outChannel = utils.InfiniteChannel()

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
	inChannel <- f
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
		f.(func())()
		d := time.Since(st)
		if d > 50*time.Millisecond {
			log.Warnf("One thread duration (%v)", time.Since(st))
		}
	}
}
