package repoview

import (
	"fmt"
	"github.com/jroimartin/gocui"
	"github.com/michael-reichenauer/gmc/common/config"
	"github.com/michael-reichenauer/gmc/utils/log"
	"github.com/michael-reichenauer/gmc/utils/ui"
)

type mainService interface {
	ToggleShowDetails()
	MainMenuItem() ui.MenuItem
	OpenRepoMenuItems() []ui.MenuItem
	RecentReposMenuItem() ui.MenuItem
}

type RepoView struct {
	view        ui.View
	ui          *ui.UI
	mainService mainService
	vm          *repoVM
}

func NewRepoView(ui *ui.UI, configService *config.Service, mainService mainService, workingFolder string) *RepoView {
	h := &RepoView{
		ui:          ui,
		mainService: mainService,
	}
	h.vm = newRepoVM(ui, h, mainService, configService, workingFolder)
	view := ui.NewViewFromPageFunc(h.viewPageData)
	view.Properties().OnLoad = h.onLoad
	view.Properties().OnClose = h.vm.close
	view.Properties().Name = "RepoView"
	view.Properties().OnMouseRight = h.vm.showContextMenu
	view.Properties().HideHorizontalScrollbar = true
	view.Properties().HasFrame = false
	h.view = view
	return h
}

func (h *RepoView) Show() {
	h.view.Show(ui.FullScreen())
	h.view.SetCurrentView()
	h.view.SetTop()
}

func (h *RepoView) NotifyChanged() {
	h.view.NotifyChanged()
}

func (h *RepoView) viewPageData(viewPort ui.ViewPage) ui.ViewText {
	repoPage, err := h.vm.GetRepoPage(viewPort)
	if err != nil {
		return ui.ViewText{Lines: []string{ui.Red(fmt.Sprintf("Error: %v", err))}}
	}

	h.setWindowTitle(repoPage.repoPath, repoPage.currentBranchName, repoPage.uncommittedChanges)

	if len(repoPage.lines) > 0 {
		//h.detailsView.SetCurrent(repoPage.currentIndex)
	}

	return ui.ViewText{Lines: repoPage.lines, Total: repoPage.total}
}

func (h *RepoView) onLoad() {
	h.view.SetKey(gocui.KeyF5, h.vm.refresh)
	h.view.SetKey(gocui.KeyEnter, h.vm.ToggleDetails)
	h.view.SetKey(gocui.KeyCtrlSpace, h.vm.commit)
	h.view.SetKey(gocui.KeyArrowRight, h.showContextMenu)
	h.view.SetKey(gocui.KeyCtrlS, h.vm.saveTotalDebugState)
	h.view.SetKey(gocui.KeyCtrlB, h.vm.ChangeBranchColor)
	h.view.SetKey(gocui.KeyCtrlD, h.vm.showDiff)
	h.view.SetKey(gocui.KeyEsc, h.ui.Quit)
	h.view.SetKey(gocui.KeyCtrlC, h.ui.Quit)
	h.view.SetKey('q', h.ui.Quit)
	h.view.SetKey(gocui.KeyCtrlQ, h.ui.Quit)

	h.vm.load()
	log.Infof("Load trigger refresh")
	h.vm.refresh()
}

func (h *RepoView) setWindowTitle(path, branch string, changes int) {
	changesText := ""
	if changes > 0 {
		changesText = fmt.Sprintf(" (*%d)", changes)
	}
	ui.SetWindowTitle(fmt.Sprintf("gmc: %s - %s%s", path, branch, changesText))
}

func (h *RepoView) showContextMenu() {
	vp := h.view.ViewPage()
	h.vm.showContextMenu(10, vp.CurrentLine-vp.FirstLine)
}

func (h *RepoView) showProgress() {

	//h.progress = h.ui.ShowProgress("Some Progress")
}
