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
	menuService *menuService
}

func NewRepoView(ui *ui.UI, configService *config.Service, mainService mainService, workingFolder string) *RepoView {
	h := &RepoView{
		ui:          ui,
		mainService: mainService,
	}
	h.vm = newRepoVM(ui, h, mainService, configService, workingFolder)
	h.menuService = newMenuService(ui, h.vm)
	h.view = h.newView()
	return h
}

func (h *RepoView) newView() ui.View {
	view := h.ui.NewViewFromPageFunc(h.viewPageData)
	view.Properties().OnLoad = h.onLoad
	view.Properties().Name = "RepoView"
	view.Properties().OnMouseRight = h.showContextMenuAt
	view.Properties().HideHorizontalScrollbar = true
	view.Properties().HasFrame = false

	view.SetKey(gocui.KeyF5, h.vm.refresh)
	view.SetKey(gocui.KeyEnter, h.vm.ToggleDetails)
	view.SetKey(gocui.KeyCtrlSpace, h.vm.commit)
	view.SetKey(gocui.KeyCtrlS, h.vm.saveTotalDebugState)
	view.SetKey(gocui.KeyCtrlB, h.vm.ChangeBranchColor)
	view.SetKey(gocui.KeyCtrlD, h.vm.showSelectedCommitDiff)
	view.SetKey(gocui.KeyEsc, h.ui.Quit)
	view.SetKey(gocui.KeyCtrlC, h.ui.Quit)
	view.SetKey('m', h.showContextMenu)
	view.SetKey('q', h.ui.Quit)
	view.SetKey(gocui.KeyCtrlQ, h.ui.Quit)

	return view
}

func (h *RepoView) Show() {
	h.view.Show(ui.FullScreen())
	h.view.SetCurrentView()
	h.view.SetTop()
}

func (h *RepoView) Close() {
	h.vm.close()
	h.view.Close()
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
	h.vm.load()
	log.Infof("Load trigger refresh")
	h.ui.PostOnUIThread(func() { h.vm.refresh() })
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
	menu := h.menuService.getContextMenu(vp.CurrentLine)
	menu.Show(11, vp.CurrentLine-vp.FirstLine)
}

func (h *RepoView) showContextMenuAt(x int, y int) {
	vp := h.view.ViewPage()
	menu := h.menuService.getContextMenu(vp.FirstLine + y)
	menu.Show(x+1, vp.CurrentLine-vp.FirstLine)
}
