package repoview

import (
	"fmt"
	"github.com/jroimartin/gocui"
	"github.com/michael-reichenauer/gmc/utils/log"
	"github.com/michael-reichenauer/gmc/utils/ui"
)

type RepoView struct {
	ui.View
	vm                 *repoVM
	repoPath           string
	isSelected         bool
	currentBranch      string
	isShowCommitStatus bool
}

func NewRepoView(uiHandler *ui.Handler, repoPath string) *RepoView {
	h := &RepoView{
		View:     uiHandler.NewView(),
		repoPath: repoPath,
		vm:       newRepoVM(repoPath),
	}
	h.Properties().OnViewData = h.onViewData
	h.Properties().OnLoad = h.onLoad
	return h
}

func (h *RepoView) onViewData(viewPort ui.ViewPort) ui.ViewData {
	repoPage, err := h.vm.GetRepoPage(viewPort.Width, viewPort.First, viewPort.Last, viewPort.Current)
	if err != nil {
		return ui.ViewData{Text: ui.Red(fmt.Sprintf("Error: %v", err)), MaxLines: 1}
	}
	log.Infof("got repo")
	h.setWindowTitle(repoPage.repoPath, repoPage.currentBranchName, repoPage.commitStatus)

	if !h.isSelected && repoPage.currentCommitIndex != -1 {
		h.isSelected = true
		//h.SetCursor(repoPage.currentCommitIndex)
	}

	return ui.ViewData{
		Text:     repoPage.text,
		MaxLines: repoPage.lines,
		First:    repoPage.first,
		Last:     repoPage.last,
		Current:  repoPage.current,
	}
}

func (h *RepoView) onLoad() {
	h.vm.Load()
	h.setWindowTitle(h.repoPath, "", "")

	h.SetKey(gocui.KeyCtrl5, gocui.ModNone, h.onRefresh)
	h.SetKey(gocui.KeyF5, gocui.ModNone, h.onRefresh)
	h.SetKey(gocui.KeyEnter, gocui.ModNone, h.onEnter)
	h.SetKey(gocui.KeyArrowLeft, gocui.ModNone, h.onLeft)
	h.SetKey(gocui.KeyArrowRight, gocui.ModNone, h.onRight)
	h.NotifyChanged()
}

func (h *RepoView) onEnter() {
	h.isShowCommitStatus = !h.isShowCommitStatus
	h.NotifyChanged()
}

func (h *RepoView) onRight() {
	h.vm.OpenBranch(h.CurrentLine())
	h.NotifyChanged()
}

func (h *RepoView) onLeft() {
	h.vm.CloseBranch(h.CurrentLine())
	h.NotifyChanged()
}

func (h *RepoView) onRefresh() {
	h.Clear()
	h.PostOnUIThread(func() {
		// Posted to allow the clear to show while new data is calculated
		h.vm.Refresh()
		h.NotifyChanged()
	})
}

func (h *RepoView) setWindowTitle(path, branch, status string) {
	statusTxt := ""
	//if h.isShowCommitStatus {
	statusTxt = fmt.Sprintf("  %s", status)
	//}
	ui.SetWindowTitle(fmt.Sprintf("gmc: %s - %s%s", path, branch, statusTxt))
}
