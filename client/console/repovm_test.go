package console

import (
	"fmt"
	"github.com/michael-reichenauer/gmc/api"
	"github.com/michael-reichenauer/gmc/utils/cui"
	"github.com/michael-reichenauer/gmc/utils/log"
	"github.com/michael-reichenauer/gmc/utils/tests"
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

func (t uiMock) NewView(text string) cui.View {
	panic("implement me")
}

func (t uiMock) NewViewFromPageFunc(f func(viewPort cui.ViewPage) cui.ViewText) cui.View {
	panic("implement me")
}

func (t uiMock) NewViewFromTextFunc(f func(viewPage cui.ViewPage) string) cui.View {
	panic("implement me")
}

func (t uiMock) ShowProgress(format string, v ...interface{}) cui.Progress {
	log.Infof("Progress %q", fmt.Sprintf(format, v...))
	return newProgressMock()
}

func (t uiMock) ShowMessageBox(title, format string, v ...interface{}) {
	log.Infof("ShowMessageBox: title: %q, msg: %q", title, fmt.Sprintf(format, v...))
}

func (t uiMock) ShowErrorMessageBox(format string, v ...interface{}) {
	log.Warnf("ShowErrorMessageBox: %q", fmt.Sprintf(format, v...))
}

func (t uiMock) ResizeAllViews() {
}

func (t uiMock) NewMenu(title string) cui.Menu {
	panic("implement me")
}

func (t uiMock) Quit() {
	t.Close()
}

type viewRepoMock struct {
	repoChanges chan api.RepoChange
}

func newViewRepoMock() api.Repo {
	return &viewRepoMock{repoChanges: make(chan api.RepoChange)}
}

func (t viewRepoMock) StartMonitor() {
}

func (t viewRepoMock) RepoChanges() chan api.RepoChange {
	return t.repoChanges
}

func (t viewRepoMock) CloseRepo() {
}

func (t viewRepoMock) TriggerRefreshModel() {
}

func (t viewRepoMock) TriggerSearch(text string) {
	panic("implement me")
}

func (t viewRepoMock) GetCommitOpenInBranches(id string) []api.Branch {
	panic("implement me")
}

func (t viewRepoMock) GetCommitOpenOutBranches(id string) []api.Branch {
	panic("implement me")
}

func (t viewRepoMock) GetCurrentNotShownBranch() (api.Branch, bool) {
	panic("implement me")
}

func (t viewRepoMock) GetCurrentBranch() (api.Branch, bool) {
	panic("implement me")
}

func (t viewRepoMock) GetLatestBranches(shown bool) []api.Branch {
	panic("implement me")
}

func (t viewRepoMock) GetAllBranches(shown bool) []api.Branch {
	panic("implement me")
}

func (t viewRepoMock) GetShownBranches(master bool) []api.Branch {
	panic("implement me")
}

func (t viewRepoMock) ShowBranch(name string) {
	panic("implement me")
}

func (t viewRepoMock) HideBranch(name string) {
	panic("implement me")
}

func (t viewRepoMock) SwitchToBranch(name string, name2 string) error {
	panic("implement me")
}

func (t viewRepoMock) PushBranch(name string) error {
	panic("implement me")
}

func (t viewRepoMock) PullBranch() error {
	panic("implement me")
}

func (t viewRepoMock) MergeBranch(name string) error {
	panic("implement me")
}

func (t viewRepoMock) CreateBranch(name string) error {
	panic("implement me")
}

func (t viewRepoMock) DeleteBranch(name string) error {
	panic("implement me")
}

func (t viewRepoMock) GetCommitDiff(id string) (api.CommitDiff, error) {
	panic("implement me")
}

func (t viewRepoMock) Commit(message string) error {
	panic("implement me")
}

func TestViewCurrent(t *testing.T) {
	tests.ManualTest(t)
	// wf:= tests.CreateTempFolder()
	// defer tests.CleanTemp()

	uim := newUIMock()
	viewer := newViewerMock(uim, func() { uim.Close() })
	viewRepo := newViewRepoMock()
	vm := newRepoVM(uim, viewer, viewRepo)
	vm.startRepoMonitor()
	defer vm.close()
	vm.triggerRefresh()
	uim.Run()
	vp, _ := vm.GetRepoPage(cui.ViewPage{Height: 10050, Width: 80})
	fmt.Printf("%s\n", strings.Join(vp.lines, "\n"))
}

func TestViewAcs(t *testing.T) {
	tests.ManualTest(t)
	// wf:= tests.CreateTempFolder()
	// defer tests.CleanTemp()

	uim := newUIMock()
	viewer := newViewerMock(uim, func() { uim.Close() })
	viewRepo := newViewRepoMock()
	vm := newRepoVM(uim, viewer, viewRepo)
	vm.startRepoMonitor()
	defer vm.close()
	vm.triggerRefresh()
	uim.Run()
	vp, _ := vm.GetRepoPage(cui.ViewPage{Height: 10050, Width: 80})
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
