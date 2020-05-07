package repoview

import (
	"fmt"
	"github.com/michael-reichenauer/gmc/common/config"
	"github.com/michael-reichenauer/gmc/utils/git"
	"github.com/michael-reichenauer/gmc/utils/log"
	"github.com/michael-reichenauer/gmc/utils/tests"
	"github.com/michael-reichenauer/gmc/utils/ui"
	"strings"
	"testing"
)

type progressMock struct {
}

func newProgressMock() *progressMock {
	return &progressMock{}
}

func (*progressMock) Close() {
}

type viewerMock struct {
	ui     *uiMock
	notify func()
}

func newViewerMock(ui *uiMock, notify func()) *viewerMock {
	if notify == nil {
		notify = func() { ui.Close() }
	}
	return &viewerMock{ui: ui, notify: notify}
}

func (t *viewerMock) NotifyChanged() {
	t.ui.PostOnUIThread(t.notify)
}

func (t *viewerMock) PostOnUIThread(f func()) {
	t.ui.PostOnUIThread(f)
}
func (t *viewerMock) Run() {
	t.ui.Run()
}

type uiMock struct {
	uiWork chan func()
}

func newUIMock() *uiMock {
	return &uiMock{uiWork: make(chan func())}
}

func (t *uiMock) Run() {
	for f := range t.uiWork {
		f()
	}
}
func (t *uiMock) Close() {
	close(t.uiWork)
}

func (t uiMock) PostOnUIThread(f func()) {
	go func() { t.uiWork <- f }()
}

func (t uiMock) NewView(text string) ui.View {
	panic("implement me")
}

func (t uiMock) NewViewFromPageFunc(f func(viewPort ui.ViewPage) ui.ViewText) ui.View {
	panic("implement me")
}

func (t uiMock) NewViewFromTextFunc(f func(viewPage ui.ViewPage) string) ui.View {
	panic("implement me")
}

func (t uiMock) ShowProgress(text string) ui.Progress {
	log.Infof("Progress %q", text)
	return newProgressMock()
}

func (t uiMock) ShowMessageBox(text string, title string) {
	log.Infof("ShowMessageBox: title: %q, msg: %q", title, text)
}

func (t uiMock) ShowErrorMessageBox(text string) {
	log.Warnf("ShowErrorMessageBox: %q", text)
}

func (t uiMock) ResizeAllViews() {
}

func (t uiMock) NewMenu(title string) ui.Menu {
	panic("implement me")
}

func (t uiMock) Quit() {
	t.Close()
}

func TestViewCurrent(t *testing.T) {
	tests.ManualTest(t)
	// wf:= tests.CreateTempFolder()
	// defer tests.CleanTemp()

	uim := newUIMock()
	viewer := newViewerMock(uim, func() { uim.Close() })
	vm := newRepoVM(uim, viewer, config.NewConfig("0.1", "", ""), git.CurrentRoot())
	vm.startRepoMonitor()
	defer vm.close()
	vm.triggerRefresh()
	uim.Run()
	vp, _ := vm.GetRepoPage(ui.ViewPage{Height: 10050, Width: 80})
	fmt.Printf("%s\n", strings.Join(vp.lines, "\n"))
}

func TestViewAcs(t *testing.T) {
	tests.ManualTest(t)
	// wf:= tests.CreateTempFolder()
	// defer tests.CleanTemp()

	uim := newUIMock()
	viewer := newViewerMock(uim, func() { uim.Close() })
	vm := newRepoVM(uim, viewer, config.NewConfig("0.1", "", ""), "C:\\Work Files\\AcmAcs")
	vm.startRepoMonitor()
	defer vm.close()
	vm.triggerRefresh()
	uim.Run()
	vp, _ := vm.GetRepoPage(ui.ViewPage{Height: 10050, Width: 80})
	fmt.Printf("%s\n", strings.Join(vp.lines, "\n"))
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
