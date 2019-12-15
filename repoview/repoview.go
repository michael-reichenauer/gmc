package repoview

import (
	"fmt"
	"github.com/jroimartin/gocui"
	"github.com/michael-reichenauer/gmc/utils/ui"
)

type RepoView struct {
	uiHandler          *ui.Handler
	viewHandler        *ui.ViewHandler
	vm                 *repoVM
	repoPath           string
	isSelected         bool
	currentBranch      string
	isShowCommitStatus bool
}

func NewRepoView(uiHandler *ui.Handler, repoPath string) *RepoView {
	return &RepoView{
		uiHandler: uiHandler,
		repoPath:  repoPath,
		vm:        newRepoVM(repoPath),
	}
}

func (h *RepoView) Properties() ui.Properties {
	return ui.Properties{
		Title:  "",
		OnLoad: h.OnLoad,
	}
}

func (h *RepoView) GetViewData(viewPort ui.ViewPort) ui.ViewData {
	repoPage, err := h.vm.GetRepoPage(viewPort.Width, viewPort.First, viewPort.Last, viewPort.Current)
	if err != nil {
		return ui.ViewData{Text: ui.Red(fmt.Sprintf("Error: %v", err)), MaxLines: 1}
	}

	h.setWindowTitle(repoPage.repoPath, repoPage.currentBranchName, repoPage.commitStatus)

	if !h.isSelected && repoPage.currentCommitIndex != -1 {
		h.isSelected = true
		h.SetCursor(repoPage.currentCommitIndex)
	}

	return ui.ViewData{
		Text:     repoPage.text,
		MaxLines: repoPage.lines,
		First:    repoPage.first,
		Last:     repoPage.last,
		Current:  repoPage.current,
	}
}

func (h *RepoView) OnLoad(view *ui.ViewHandler) {
	h.viewHandler = view

	h.vm.Load()
	h.setWindowTitle(h.repoPath, "", "")

	h.viewHandler.SetKey(gocui.KeyCtrl5, gocui.ModNone, h.onRefresh)
	h.viewHandler.SetKey(gocui.KeyF5, gocui.ModNone, h.onRefresh)
	h.viewHandler.SetKey(gocui.KeyEnter, gocui.ModNone, h.onEnter)
	h.viewHandler.SetKey(gocui.KeyArrowLeft, gocui.ModNone, h.onLeft)
	h.viewHandler.SetKey(gocui.KeyArrowRight, gocui.ModNone, h.onRight)
	h.viewHandler.NotifyChanged()
}

func (h *RepoView) onEnter() {
	h.isShowCommitStatus = !h.isShowCommitStatus
	h.viewHandler.NotifyChanged()
}

func (h *RepoView) onRight() {
	h.vm.OpenBranch(h.viewHandler.CurrentLine)
	h.viewHandler.NotifyChanged()
}

func (h *RepoView) onLeft() {
	h.vm.CloseBranch(h.viewHandler.CurrentLine)
	h.viewHandler.NotifyChanged()
}

func (h *RepoView) onRefresh() {
	h.viewHandler.Clear()
	h.viewHandler.RunOnUI(func() {
		h.vm.Refresh()
		h.viewHandler.NotifyChanged()
	})
}

func (h *RepoView) setWindowTitle(path, branch, status string) {
	statusTxt := ""
	//if h.isShowCommitStatus {
	statusTxt = fmt.Sprintf("  %s", status)
	//}
	ui.SetWindowTitle(fmt.Sprintf("gmc: %s - %s%s", path, branch, statusTxt))
}

func (h *RepoView) SetCursor(line int) {
	//	h.setCursor(g, view, line)
}
