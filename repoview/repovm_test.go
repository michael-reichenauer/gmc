package repoview

import (
	"github.com/michael-reichenauer/gmc/utils/git"
	"github.com/michael-reichenauer/gmc/utils/tests"
	"testing"
)

type viewerMock struct {
	uiWork chan func()
}

func newViewerMock() *viewerMock {
	return &viewerMock{uiWork: make(chan func())}
}

func (t *viewerMock) WaitNotifyChanged() {
	for f := range t.uiWork {
		f()
	}
}

func (t *viewerMock) NotifyChanged() {
	close(t.uiWork)
}

func (t *viewerMock) PostOnUIThread(f func()) {
	t.uiWork <- f
}

func TestViewCurrent(t *testing.T) {
	tests.ManualTest(t)
	viewer := newViewerMock()
	vm := newRepoVM(ui, viewer, ms, cs, git.CurrentRoot())
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
