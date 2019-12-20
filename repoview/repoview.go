package repoview

import (
	"fmt"
	"github.com/jroimartin/gocui"
	"github.com/michael-reichenauer/gmc/repoview/model"
	"github.com/michael-reichenauer/gmc/utils/ui"
)

type RepoView struct {
	ui.View
	vm            *repoVM
	repoPath      string
	isSelected    bool
	currentBranch string
}

func newRepoView(uiHandler *ui.UI, model *model.Model) *RepoView {
	h := &RepoView{
		vm: newRepoVM(model),
	}
	h.View = uiHandler.NewView(h.viewData)
	h.Properties().OnLoad = h.onLoad
	return h
}

func (h *RepoView) viewData(viewPort ui.ViewPort) ui.ViewData {
	repoPage, err := h.vm.GetRepoPage(viewPort.Width, viewPort.First, viewPort.Last, viewPort.Current)
	if err != nil {
		return ui.ViewData{Text: ui.Red(fmt.Sprintf("Error: %v", err)), MaxLines: 1}
	}

	h.setWindowTitle(repoPage.repoPath, repoPage.currentBranchName, repoPage.uncommittedChanges)

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
	h.setWindowTitle(h.repoPath, "", 0)

	h.SetKey(gocui.KeyCtrl5, gocui.ModNone, h.onRefresh)
	h.SetKey(gocui.KeyF5, gocui.ModNone, h.onRefresh)
	h.SetKey(gocui.KeyEnter, gocui.ModNone, h.onEnter)
	h.SetKey(gocui.KeyArrowLeft, gocui.ModNone, h.onLeft)
	h.SetKey(gocui.KeyArrowRight, gocui.ModNone, h.onRight)
	h.NotifyChanged()
}

func (h *RepoView) onEnter() {
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

func (h *RepoView) setWindowTitle(path, branch string, changes int) {
	changesText := ""
	if changes > 0 {
		changesText = fmt.Sprintf(" {%d}", changes)
	}
	ui.SetWindowTitle(fmt.Sprintf("gmc: %s%s - %s", path, changesText, branch))
}
