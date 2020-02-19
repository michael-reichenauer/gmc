package repoview

import (
	"fmt"
	"github.com/jroimartin/gocui"
	"github.com/michael-reichenauer/gmc/repoview/viewmodel"
	"github.com/michael-reichenauer/gmc/utils/ui"
)

type mainController interface {
	ToggleDetails()
	MainMenuItem() ui.MenuItem
	OpenRepoMenuItems() []ui.MenuItem
	RecentReposMenuItem() ui.MenuItem
	ShowDiff(index int)
	HideDiff()
}

type RepoView struct {
	ui.View
	uiHandler    *ui.UI
	main         mainController
	detailsView  *DetailsView
	vm           *repoVM
	emptyMessage string
}

func NewRepoView(uiHandler *ui.UI, model *viewmodel.Service, detailsView *DetailsView, mainController mainController) *RepoView {
	h := &RepoView{
		uiHandler:   uiHandler,
		detailsView: detailsView,
		main:        mainController,
	}
	h.View = uiHandler.NewViewFromPageFunc(h.viewData)
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
	} else if h.emptyMessage != "" {
		return ui.ViewData{Lines: []string{h.emptyMessage}}
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
	h.SetKey(gocui.KeyCtrlD, gocui.ModNone, h.showDiff)
	h.SetKey(gocui.KeyEsc, gocui.ModNone, h.uiHandler.Quit)
	h.SetKey(gocui.KeyCtrlC, gocui.ModNone, h.uiHandler.Quit)
	h.SetKey('q', gocui.ModNone, h.uiHandler.Quit)
	h.SetKey(gocui.KeyCtrlQ, gocui.ModNone, h.uiHandler.Quit)
	h.NotifyChanged()
}

func (h *RepoView) onEnter() {
	h.main.ToggleDetails()
	h.vm.ToggleDetails()
}

func (h *RepoView) onRight() {
	menu := ui.NewMenu(h.uiHandler, "Show")
	items := h.vm.GetOpenBranchMenuItems(h.ViewPage().CurrentLine)
	menu.AddItems(items)

	menu.Add(ui.SeparatorMenuItem)
	menu.Add(ui.MenuItem{Text: "Commit Diff ...", Key: "Ctrl-D", Action: func() {
		h.main.ShowDiff(h.ViewPage().CurrentLine)
	}})
	switchItems := h.vm.GetSwitchBranchMenuItems()
	menu.Add(ui.MenuItem{Text: "Switch/Checkout", SubItems: switchItems})
	menu.Add(h.main.RecentReposMenuItem())
	menu.Add(h.main.MainMenuItem())

	y := h.ViewPage().CurrentLine - h.ViewPage().FirstLine + 2
	menu.Show(10, y)
}

func (h *RepoView) onLeft() {
	menu := ui.NewMenu(h.uiHandler, "Hide Branch")
	items := h.vm.GetCloseBranchMenuItems()
	if len(items) == 0 {
		return
	}
	menu.AddItems(items)

	y := h.ViewPage().CurrentLine - h.ViewPage().FirstLine + 2
	menu.Show(10, y)
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

func (h *RepoView) SetEmptyMessage(message string) {
	h.emptyMessage = message
}

func (h *RepoView) showDiff() {
	h.main.ShowDiff(h.ViewPage().CurrentLine)
}
