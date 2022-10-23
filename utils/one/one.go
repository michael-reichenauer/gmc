package one

import (
	"fmt"
	"time"

	"github.com/michael-reichenauer/gmc/utils"
	"github.com/michael-reichenauer/gmc/utils/log"
)

var inChannel, outChannel = utils.InfiniteChannel()

func Run() {
	for f := range outChannel {
		if f == nil {
			break
		}
		f.(func())()
	}
}

func RunWith(otherFunc func(doFunc func())) {
	for f := range outChannel {
		if f == nil {
			break
		}
		otherFunc(f.(func()))
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
