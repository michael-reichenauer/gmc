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
	ui          ui.UI
	mainService mainService
	vm          *repoVM
	menuService *menuService
	searchView  *SearchView
}

func (t *RepoView) SetCurrentView() {
	t.view.SetCurrentView()
	t.view.NotifyChanged()
}

func (t *RepoView) ScrollVertical(scroll int) {
	t.view.ScrollVertical(scroll)

}

func NewRepoView(ui ui.UI, configService *config.Service, mainService mainService, workingFolder string) *RepoView {
	h := &RepoView{
		ui:          ui,
		mainService: mainService,
	}
	h.vm = newRepoVM(ui, h, mainService, configService, workingFolder)
	h.menuService = newMenuService(ui, h.vm)
	h.view = h.newView()
	return h
}

func (t *RepoView) newView() ui.View {
	view := t.ui.NewViewFromPageFunc(t.viewPageData)
	view.Properties().OnLoad = t.onLoad
	view.Properties().Name = "RepoView"
	view.Properties().OnMouseLeft = t.mouseLeft
	view.Properties().OnMouseRight = t.showContextMenuAt
	view.Properties().HideHorizontalScrollbar = true
	view.Properties().HasFrame = false

	view.SetKey(gocui.KeyF5, t.vm.triggerRefresh)
	view.SetKey(gocui.KeyCtrlD, t.vm.showSelectedCommitDiff)
	//view.SetKey(gocui.KeyEnter, h.vm.ToggleDetails)
	view.SetKey(gocui.KeyCtrlS, t.vm.showCommitDialog)
	view.SetKey(gocui.KeyCtrlB, t.vm.showCreateBranchDialog)
	view.SetKey(gocui.KeyCtrlP, t.vm.PushCurrentBranch)
	view.SetKey(gocui.KeyCtrlU, t.vm.PullCurrentBranch)
	view.SetKey(gocui.KeyCtrlF, t.showSearchView)
	view.SetKey(gocui.KeyTab, t.nextView)
	view.SetKey('f', t.showSearchView)
	//view.SetKey(gocui.KeyCtrlS, h.vm.saveTotalDebugState)
	//view.SetKey(gocui.KeyCtrlB, h.vm.ChangeBranchColor)

	view.SetKey('m', t.showContextMenu)
	view.SetKey(gocui.KeyEsc, t.quit)
	view.SetKey(gocui.KeyCtrlC, t.ui.Quit)

	return view
}

func (t *RepoView) Show() {
	t.view.Show(ui.FullScreen())
	t.view.SetCurrentView()
	t.view.SetTop()
}

func (t *RepoView) Close() {
	t.vm.close()
	t.view.Close()
}

func (t *RepoView) NotifyChanged() {
	t.view.NotifyChanged()
}

func (t *RepoView) viewPageData(viewPort ui.ViewPage) ui.ViewText {
	repoPage, err := t.vm.GetRepoPage(viewPort)
	if err != nil {
		return ui.ViewText{Lines: []string{ui.Red(fmt.Sprintf("Error: %v", err))}}
	}

	t.setWindowTitle(repoPage)

	if len(repoPage.lines) > 0 {
		//h.detailsView.SetCurrent(repoPage.currentIndex)
	}

	return ui.ViewText{Lines: repoPage.lines, Total: repoPage.total}
}

func (t *RepoView) onLoad() {
	t.vm.startRepoMonitor()
	log.Infof("Load trigger refresh")
	t.vm.triggerRefresh()
}

func (t *RepoView) setWindowTitle(port repoPage) {
	changesText := ""
	if port.uncommittedChanges > 0 {
		changesText = fmt.Sprintf(" (*%d)", port.uncommittedChanges)
	}
	ui.SetWindowTitle(fmt.Sprintf("gmc: %s - %s%s   (%s)",
		port.repoPath, port.currentBranchName, changesText, port.selectedBranchName))
}

func (t *RepoView) showContextMenu() {
	vp := t.view.ViewPage()
	menu := t.menuService.getContextMenu(vp.CurrentLine)
	menu.Show(11, vp.CurrentLine-vp.FirstLine)
}

func (t *RepoView) showContextMenuAt(x int, y int) {
	vp := t.view.ViewPage()
	menu := t.menuService.getContextMenu(vp.FirstLine + y)
	menu.Show(x+1, vp.CurrentLine-vp.FirstLine)
}

func (t *RepoView) mouseLeft(x int, y int) {
	vp := t.view.ViewPage()
	selectedLine := vp.FirstLine + y
	t.view.SetCurrentLine(selectedLine)
	if !t.vm.isMoreClick(x, y) {
		return
	}

	menu := t.menuService.getShowMoreMenu(selectedLine)
	menu.Show(x+3, y+2)
}

func (t *RepoView) showSearchView() {
	if t.searchView != nil {
		return
	}

	mb := ui.Relative(ui.FullScreen(), func(b ui.Rect) ui.Rect {
		return ui.Rect{X: b.X, Y: b.Y + 2, W: b.W, H: b.H - 2}
	})
	t.view.SetBound(mb)

	t.searchView = NewSearchView(t.ui, t)
	t.searchView.Show()
}

func (t *RepoView) Search(text string) {
	log.Infof("Search in search %q", text)
	t.vm.SetSearch(text)
}

func (t *RepoView) CloseSearch() {
	if t.searchView != nil {
		t.searchView = nil
	}
	t.vm.SetSearch("")
	t.view.SetBound(ui.FullScreen())
}

func (t *RepoView) nextView() {
	if t.searchView != nil {
		t.searchView.SetCurrentView()
	}
}

func (t *RepoView) quit() {
	if t.searchView != nil {
		t.searchView.SetCurrentView()
		return
	}
	t.ui.Quit()
}
