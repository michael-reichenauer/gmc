package console

import (
	"fmt"

	"github.com/jroimartin/gocui"
	"github.com/michael-reichenauer/gmc/api"
	"github.com/michael-reichenauer/gmc/utils/cui"
	"github.com/michael-reichenauer/gmc/utils/log"
)

type mainService interface {
	OpenRepoMenuItems() []cui.MenuItem
}

type RepoView struct {
	view        cui.View
	ui          cui.UI
	mainService mainService
	vm          *repoVM
	menuService Menus
	searchView  *SearchView
	detailsView *DetailsView
}

func NewRepoView(ui cui.UI, api api.Api, mainService mainService, repoID string) *RepoView {
	h := &RepoView{
		ui:          ui,
		mainService: mainService,
	}
	h.vm = newRepoVM(ui, h, api, repoID)
	h.menuService = newMenus(ui, h.vm)
	h.view = h.newView()
	return h
}

func (t *RepoView) SetCurrentView() {
	t.view.SetCurrentView()
	t.view.NotifyChanged()
}

func (t *RepoView) ScrollVertical(scroll int) {
	t.view.ScrollVertical(scroll)
}

func (t *RepoView) OpenRepoMenuItems() []cui.MenuItem {
	return t.mainService.OpenRepoMenuItems()
}

func (t *RepoView) newView() cui.View {
	view := t.ui.NewViewFromPageFunc(t.viewPageData)
	view.Properties().OnLoad = t.onLoad
	view.Properties().Name = "RepoView"
	view.Properties().OnMouseLeft = t.mouseLeft
	view.Properties().OnMouseRight = t.showContextMenuAt
	view.Properties().OnMoved = t.onMoved
	view.Properties().HideHorizontalScrollbar = true
	view.Properties().HasFrame = false

	view.SetKey(gocui.KeyEnter, t.onEnterClick)
	view.SetKey(gocui.KeyTab, t.onTabClick)
	view.SetKey(gocui.KeyF5, t.vm.triggerRefresh)
	view.SetKey('r', t.vm.triggerRefresh)
	view.SetKey('R', t.vm.triggerRefresh)
	view.SetKey(gocui.KeyCtrlR, t.vm.triggerRefresh)

	view.SetKey('d', t.vm.showSelectedCommitDiff)
	view.SetKey('D', t.vm.showSelectedCommitDiff)
	view.SetKey(gocui.KeyCtrlD, t.vm.showSelectedCommitDiff)
	view.SetKey('c', t.vm.showCommitDialog)
	view.SetKey('C', t.vm.showCommitDialog)

	view.SetKey('b', t.vm.showCreateBranchDialog)
	view.SetKey('B', t.vm.showCreateBranchDialog)
	view.SetKey('p', t.vm.PushCurrentBranch)
	view.SetKey('P', t.vm.PushCurrentBranch)
	view.SetKey('u', t.vm.PullCurrentBranch)
	view.SetKey('U', t.vm.PullCurrentBranch)
	view.SetKey('m', t.showContextMenu)
	view.SetKey('M', t.showContextMenu)

	view.SetKey('f', t.vm.ShowSearchView)
	view.SetKey('F', t.vm.ShowSearchView)
	view.SetKey('a', t.showAbout)
	view.SetKey('h', func() { ShowHelpDlg(t.ui) })
	view.SetKey('H', func() { ShowHelpDlg(t.ui) })

	view.SetKey(gocui.KeyArrowRight, t.showCommitBranchesMenu)
	view.SetKey(gocui.KeyArrowLeft, t.showHideBranchesMenu)

	view.SetKey(gocui.KeyEsc, t.onEscKey)
	view.SetKey(gocui.KeyCtrlC, t.ui.Quit)

	return view
}

func (t *RepoView) Show() {
	t.view.Show(cui.FullScreen())
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

func (t *RepoView) ShowLineAtTop(line int) {
	t.view.ShowLineAtTop(line, false)
}

func (t *RepoView) viewPageData(viewPort cui.ViewPage) cui.ViewText {
	repoPage, err := t.vm.GetRepoPage(viewPort)
	if err != nil {
		return cui.ViewText{Lines: []string{cui.Red(fmt.Sprintf("Error: %v", err))}}
	}

	t.setWindowTitle(repoPage)

	return cui.ViewText{Lines: repoPage.lines, Total: repoPage.total}
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
	cui.SetWindowTitle(fmt.Sprintf("gmc: %s - %s%s   (%s)",
		port.repoPath, port.currentBranchName, changesText, port.selectedBranchName))
}

// Called by left-arrow, to show a hide branches menu
func (t *RepoView) showHideBranchesMenu() {
	if t.isInSearchMode() {
		return
	}
	vp := t.view.ViewPage()
	line := vp.CurrentLine

	menu := t.menuService.GetHideBranchesMenu()
	menu.Show(11, line-vp.FirstLine)
}

// Called by right-arrow to show commit branches to show/expand
func (t *RepoView) showCommitBranchesMenu() {
	if t.isInSearchMode() {
		return
	}
	vp := t.view.ViewPage()
	line := vp.CurrentLine

	menu := t.menuService.GetShowBranchesMenu(line)
	menu.Show(11, line-vp.FirstLine)
}

func (t *RepoView) onEnterClick() {
	if t.isInSearchMode() {
		c := t.vm.repo.Commits[t.vm.currentIndex]
		b := t.vm.repo.Branches[c.BranchIndex]
		t.searchView.onCancel()
		t.ui.Post(func() {
			t.vm.ShowBranch(b.Name, c.ID)
		})

		return
	}

	if t.isDetailsMode() {
		t.hideCommitDetails()
		return
	}

	t.ShowCommitDetails()
}

func (t *RepoView) onTabClick() {
	if t.isDetailsMode() {
		t.detailsView.SetCurrentView()
	}
}

func (t *RepoView) showContextMenu() {
	// Show context menu
	vp := t.view.ViewPage()
	menu := t.menuService.GetMainMenu(vp.CurrentLine)
	menu.Show(40, 0)
}

func (t *RepoView) showAbout() {
	ShowAboutDlg(t.ui)
}

func (t *RepoView) showContextMenuAt(x int, y int) {
	if t.isInSearchMode() {
		return
	}
	vp := t.view.ViewPage()
	menu := t.menuService.GetMainMenu(vp.FirstLine + y)
	menu.Show(x+1, vp.CurrentLine-vp.FirstLine)
}

func (t *RepoView) mouseLeft(x int, y int) {
	if t.isInSearchMode() {
		return
	}
	vp := t.view.ViewPage()
	selectedLine := vp.FirstLine + y
	t.view.SetCurrentLine(selectedLine)
	t.ui.Post(func() {
		if t.vm.isGraphClick(x, y) {
			menu := t.menuService.GetShowBranchesMenu(selectedLine)
			menu.Show(x+3, y+2)
		}
	})
}

func (t *RepoView) ShowSearchView() {
	if t.isInSearchMode() {
		return
	}

	mb := cui.Relative(cui.FullScreen(), func(b cui.Rect) cui.Rect {
		return cui.Rect{X: b.X, Y: b.Y + 2, W: b.W, H: b.H - 2}
	})
	t.view.SetBound(mb)

	t.searchView = NewSearchView(t.ui, t)
	t.searchView.Show()
}

func (t *RepoView) onMoved() {
	if !t.isDetailsMode() {
		return
	}

	vp := t.view.ViewPage()
	line := vp.CurrentLine
	t.detailsView.SetCurrentLine(line, t.vm.repo, t.vm.repoID, t.vm.api)
}

func (t *RepoView) ShowCommitDetails() {
	if t.isDetailsMode() {
		t.hideCommitDetails()
		return
	}

	hight := 15
	mb := cui.Relative(cui.FullScreen(), func(b cui.Rect) cui.Rect {
		return cui.Rect{X: b.X, Y: b.Y, W: b.W, H: b.H - hight}
	})
	t.view.SetBound(mb)

	t.detailsView = NewDetailsView(t.ui, t)
	detailsBounds := cui.Relative(mb, func(b cui.Rect) cui.Rect {
		return cui.Rect{X: b.X, Y: b.Y + b.H + 1, W: b.W - 1, H: hight - 1}
	})

	t.detailsView.Show(detailsBounds)
	vp := t.view.ViewPage()
	line := vp.CurrentLine
	t.detailsView.SetCurrentLine(line, t.vm.repo, t.vm.repoID, t.vm.api)
}

func (t *RepoView) hideCommitDetails() {
	if !t.isDetailsMode() {
		return
	}

	t.detailsView.Close()
	t.detailsView = nil
	t.view.SetBound(cui.FullScreen())
}

func (t *RepoView) Search(text string) {
	log.Infof("Search in search %q", text)
	t.vm.SetSearch(text)
}

func (t *RepoView) CloseSearch() {
	if t.isInSearchMode() {
		t.searchView = nil
	}
	t.vm.SetSearch("")
	t.view.SetBound(cui.FullScreen())
}

func (t *RepoView) onEscKey() {
	if t.isInSearchMode() {
		t.searchView.SetCurrentView()
		return
	}
	t.ui.Quit()
}

func (t *RepoView) isInSearchMode() bool {
	return t.searchView != nil
}

func (t *RepoView) isDetailsMode() bool {
	return t.detailsView != nil
}
