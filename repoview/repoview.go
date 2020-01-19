package repoview

import (
	"fmt"
	"github.com/jroimartin/gocui"
	"github.com/michael-reichenauer/gmc/repoview/viewmodel"
	"github.com/michael-reichenauer/gmc/utils/ui"
)

//
type RepoView struct {
	ui.View
	uiHandler   *ui.UI
	main        mainController
	detailsView *DetailsView
	vm          *repoVM
}

func newRepoView(uiHandler *ui.UI, model *viewmodel.Service, detailsView *DetailsView, mainController mainController) *RepoView {
	h := &RepoView{
		uiHandler:   uiHandler,
		detailsView: detailsView,
		main:        mainController,
	}
	h.View = uiHandler.NewView(h.viewData)
	h.Properties().OnLoad = h.onLoad
	h.Properties().Name = "RepoView"
	h.vm = newRepoVM(model, h)
	return h
}

func (h *RepoView) viewData(viewPort ui.ViewPage) ui.ViewData {
	// log.Infof("repo viewData ...")
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

	// log.Infof("repo view data %d lines", len(repoPage.lines))
	return ui.ViewData{Lines: repoPage.lines, FirstIndex: repoPage.firstIndex, Total: repoPage.total}
}

func (h *RepoView) onLoad() {
	h.vm.onLoad()
	h.setWindowTitle("", "", 0)

	h.SetKey(gocui.KeyCtrl5, gocui.ModNone, h.onRefresh)
	h.SetKey(gocui.KeyF5, gocui.ModNone, h.onRefresh)
	h.SetKey(gocui.KeyEnter, gocui.ModNone, h.onEnter)
	h.SetKey(gocui.KeyArrowLeft, gocui.ModNone, h.onLeft)
	h.SetKey(gocui.KeyArrowRight, gocui.ModNone, h.onRight)
	h.SetKey(gocui.KeyCtrlS, gocui.ModNone, h.onTrace)
	h.SetKey(gocui.KeyCtrlB, gocui.ModNone, h.onBranchColor)
	h.SetKey(gocui.KeyEsc, gocui.ModNone, h.uiHandler.Quit)
	h.SetKey('m', gocui.ModNone, h.onMenu)
	h.NotifyChanged()
}

func (h *RepoView) onEnter() {
	h.main.ToggleDetails()
	h.vm.ToggleDetails()
}

func (h *RepoView) onRight() {
	items := h.vm.GetOpenBranchItems(h.ViewPage().CurrentLine)
	y := h.ViewPage().CurrentLine - h.ViewPage().FirstLine
	menu := ui.NewMenu(h.uiHandler)
	menu.AddItems(items)
	menu.Show(10, y)
}

func (h *RepoView) onLeft() {
	h.vm.CloseBranch(h.ViewPage().CurrentLine)
	h.NotifyChanged()
}

func (h *RepoView) onRefresh() {
	h.Clear()
	h.PostOnUIThread(func() {
		// Posted to allow the clear to show while new data is calculated
		h.vm.Refresh()
	})
}

func (h *RepoView) setWindowTitle(path, branch string, changes int) {
	changesText := ""
	if changes > 0 {
		changesText = fmt.Sprintf(" (*%d)", changes)
	}
	ui.SetWindowTitle(fmt.Sprintf("gmc: %s - %s%s", path, branch, changesText))
}

func (h *RepoView) onTrace() {
	h.Clear()
	h.PostOnUIThread(func() {
		// Posted to allow the clear to show while new data is calculated
		h.vm.RefreshTrace(h.ViewPage())
		h.NotifyChanged()
	})
}

func (h *RepoView) onBranchColor() {
	h.vm.ChangeBranchColor(h.ViewPage().CurrentLine)
	h.NotifyChanged()
}

func (h *RepoView) onMenu() {
	menu := ui.NewMenu(h.uiHandler)
	menu.Add(
		ui.MenuItem{Text: "About", Action: h.main.ShowAbout},
	)
	menu.Show(10, 5)
}
