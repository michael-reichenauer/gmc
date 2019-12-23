package repoview

import (
	"fmt"
	"github.com/jroimartin/gocui"
	"github.com/michael-reichenauer/gmc/repoview/viewmodel"
	"github.com/michael-reichenauer/gmc/utils/log"
	"github.com/michael-reichenauer/gmc/utils/ui"
)

//
type RepoView struct {
	ui.View
	detailsView *DetailsView
	vm          *repoVM
}

func newRepoView(uiHandler *ui.UI, model *viewmodel.Model, detailsView *DetailsView) *RepoView {
	h := &RepoView{
		detailsView: detailsView,
	}
	h.View = uiHandler.NewView(h.viewData)
	h.Properties().OnLoad = h.onLoad
	h.vm = newRepoVM(model, h)
	return h
}

func (h *RepoView) viewData(viewPort ui.ViewPort) ui.ViewData {
	log.Infof("repo viewData ...")
	repoPage, err := h.vm.GetRepoPage(viewPort)
	if err != nil {
		return ui.ViewData{Lines: []string{ui.Red(fmt.Sprintf("Error: %v", err))}}
	}

	h.setWindowTitle(repoPage.repoPath, repoPage.currentBranchName, repoPage.uncommittedChanges)

	//if !h.isSelected && repoPage.currentCommitIndex != -1 {
	//	h.isSelected = true
	//	//h.SetCursor(repoPage.currentCommitIndex)
	//}
	if len(repoPage.lines) > 0 {
		h.detailsView.SetCurrent(repoPage.currentIndex)
	} else {
		return ui.ViewData{Lines: []string{"  Reading repo, please wait ..."}}
	}

	log.Infof("repo view data %d lines", len(repoPage.lines))
	return ui.ViewData{Lines: repoPage.lines, FirstIndex: repoPage.firstIndex, Total: repoPage.total}
}

func (h *RepoView) onLoad() {
	log.Infof("repo onload ...")
	h.vm.Load()
	h.setWindowTitle("", "", 0)

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
