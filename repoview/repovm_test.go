package repoview

import (
	"github.com/michael-reichenauer/gmc/utils/tests"
	"testing"
)

type mock struct {
	uiWork chan func()
}

func (m *mock) NotifyChanged() {
	close(m.uiWork)
}
func (m *mock) PostOnUIThread(f func()) {
	select {
	case m.uiWork <- f:
	default:
	}
}

func TestViewCurrent(t *testing.T) {
	tests.ManualTest(t)
	// cs := config.NewConfig("0.0", "")
	// m := &mock{uiWork: make(chan func())}
	// vm := newRepoVM(m, nil, cs, "")
	// vm.startRepoMonitor()
	// vm.triggerRefresh()
	// for f := range m.uiWork {
	// 	f()
	// }
	// vd, _ := vm.GetRepoPage(ui.ViewPage{Height: 20, Width: 120})
	// fmt.Printf("%s\n", strings.Join(vd.lines, "\n"))
}

func TestViewAcst(t *testing.T) {
	tests.ManualTest(t)
	// cs := config.NewConfig("0.0", "c:/code/AcmAcs")
	// m := &mock{uiWork: make(chan func())}
	// vm := newRepoVM(m, nil, cs, "c:/code/AcmAcs")
	// vm.startRepoMonitor()
	// vm.triggerRefresh()
	// for f := range m.uiWork {
	// 	f()
	// }
	// vd, _ := vm.GetRepoPage(ui.ViewPage{Height: 20, Width: 120})
	// fmt.Printf("%s\n", strings.Join(vd.lines, "\n"))
}
