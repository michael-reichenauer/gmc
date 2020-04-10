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
	ShowDiff(diffGetter DiffGetter, commitID string)
	HideDiff()
	HideCommit()
	NewMenu(title string) *ui.Menu
	Commit(committer Committer)
}

type RepoView struct {
	ui.View
	uiHandler   *ui.UI
	mainService mainService
	vm          *repoVM
}

func NewRepoView(uiHandler *ui.UI, configService *config.Service, mainService mainService, workingFolder string) *RepoView {
	h := &RepoView{
		uiHandler:   uiHandler,
		mainService: mainService,
	}
	h.vm = newRepoVM(h, mainService, configService, workingFolder)
	h.View = uiHandler.NewViewFromPageFunc(h.viewPageData)
	h.Properties().OnLoad = h.onLoad
	h.Properties().OnClose = h.vm.close
	h.Properties().Name = "RepoView"
	h.Properties().OnMouseRight = h.vm.showContextMenu
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
	h.SetKey(gocui.KeyEsc, h.uiHandler.Quit)
	h.SetKey(gocui.KeyCtrlC, h.uiHandler.Quit)
	h.SetKey('q', h.uiHandler.Quit)
	h.SetKey(gocui.KeyCtrlQ, h.uiHandler.Quit)

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
