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
	ui.View
	ui          *ui.UI
	mainService mainService
	vm          *repoVM
	progress    *ui.Progress
}

func NewRepoView(ui *ui.UI, configService *config.Service, mainService mainService, workingFolder string) *RepoView {
	h := &RepoView{
		ui:          ui,
		mainService: mainService,
	}
	h.vm = newRepoVM(ui, h, mainService, configService, workingFolder)
	h.View = ui.NewViewFromPageFunc(h.viewPageData)
	h.Properties().OnLoad = h.onLoad
	h.Properties().OnClose = h.vm.close
	h.Properties().Name = "RepoView"
	h.Properties().OnMouseRight = h.vm.showContextMenu
	h.Properties().HideHorizontalScrollbar = true
	return h
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
	h.SetKey(gocui.KeyF5, h.vm.refresh)
	h.SetKey(gocui.KeyEnter, h.vm.ToggleDetails)
	h.SetKey(gocui.KeyCtrlSpace, h.vm.commit)
	h.SetKey(gocui.KeyArrowRight, h.showContextMenu)
	h.SetKey(gocui.KeyCtrlS, h.vm.saveTotalDebugState)
	h.SetKey(gocui.KeyCtrlB, h.vm.ChangeBranchColor)
	h.SetKey(gocui.KeyCtrlD, h.vm.showDiff)
	h.SetKey(gocui.KeyEsc, h.ui.Quit)
	h.SetKey(gocui.KeyCtrlC, h.ui.Quit)
	h.SetKey('q', h.ui.Quit)
	h.SetKey(gocui.KeyCtrlQ, h.ui.Quit)

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
	p := h.ViewPage()
	h.vm.showContextMenu(10, p.CurrentLine-p.FirstLine)
}

func (h *RepoView) showProgress() {

	h.progress = h.ui.ShowProgress("Some Progress")
}
