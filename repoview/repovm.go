package repoview

import (
	"context"
	"fmt"
	"github.com/michael-reichenauer/gmc/common/config"
	"github.com/michael-reichenauer/gmc/repoview/viewmodel"
	"github.com/michael-reichenauer/gmc/utils/git"
	"github.com/michael-reichenauer/gmc/utils/log"
	"github.com/michael-reichenauer/gmc/utils/ui"
	"github.com/thoas/go-funk"
)

type repoPage struct {
	lines              []string
	total              int
	repoPath           string
	currentBranchName  string
	uncommittedChanges int
}

type repoVM struct {
	ui               *ui.UI
	repoViewer       ui.Viewer
	mainService      mainService
	viewModelService *viewmodel.Service
	repoLayout       *repoLayout
	isDetails        bool
	workingFolder    string
	cancel           context.CancelFunc
	repo             viewmodel.ViewRepo
	firstIndex       int
	currentIndex     int
	isLoading        bool
}

type trace struct {
	RepoPath    string
	ViewPage    ui.ViewPage
	BranchNames []string
}

func newRepoVM(
	ui *ui.UI,
	repoViewer ui.Viewer,
	mainService mainService,
	configService *config.Service,
	workingFolder string) *repoVM {
	viewModelService := viewmodel.NewService(configService, workingFolder)
	return &repoVM{
		ui:               ui,
		repoViewer:       repoViewer,
		mainService:      mainService,
		viewModelService: viewModelService,
		repoLayout:       newRepoLayout(viewModelService),
		workingFolder:    workingFolder,
	}
}

func (h *repoVM) load() {
	h.isLoading = true
	ctx, cancel := context.WithCancel(context.Background())
	h.cancel = cancel
	go h.monitorModelRoutine(ctx)
}

func (h *repoVM) close() {
	h.cancel()
}

func (h *repoVM) monitorModelRoutine(ctx context.Context) {
	h.viewModelService.StartMonitor(ctx)

	for vr := range h.viewModelService.ViewRepos {
		log.Infof("Detected model change")
		h.repoViewer.PostOnUIThread(func() {
			h.isLoading = false
			h.repo = vr
			h.repoViewer.NotifyChanged()
		})
	}
}

func (h *repoVM) GetRepoPage(viewPage ui.ViewPage) (repoPage, error) {
	//t := timer.Start()
	//defer log.Infof("GetRepoPage %d %d %v", viewPage.FirstLine, viewPage.CurrentLine, t)
	if h.isLoading {
		return repoPage{
			repoPath: h.repo.RepoPath,
			lines:    []string{fmt.Sprintf("Loading %s ...", h.workingFolder)},
			total:    1,
		}, nil
	}

	firstIndex, lines := h.getLines(viewPage)
	h.firstIndex = firstIndex
	h.currentIndex = viewPage.CurrentLine
	return repoPage{
		repoPath:           h.repo.RepoPath,
		lines:              lines,
		total:              len(h.repo.Commits),
		uncommittedChanges: h.repo.UncommittedChanges,
		currentBranchName:  h.repo.CurrentBranchName,
	}, nil
}

func (h *repoVM) getLines(viewPage ui.ViewPage) (int, []string) {
	firstIndex, commits := h.getCommits(viewPage)
	// var currentLineCommit viewmodel.Commit
	// if len(commits) > 0 {
	// 	currentLineCommit = commits[viewPage.CurrentLine-viewPage.FirstLine]
	// }

	return firstIndex, h.repoLayout.getPageLines(commits, viewPage.Width, "")
}

func (h *repoVM) getCommits(viewPage ui.ViewPage) (int, []viewmodel.Commit) {
	firstIndex := viewPage.FirstLine
	count := viewPage.Height
	if count > len(h.repo.Commits) {
		// Requested count larger than available, return just all available commits
		count = len(h.repo.Commits)
	}

	if firstIndex+count >= len(h.repo.Commits) {
		// Requested commits past available, adjust to return available commits
		firstIndex = len(h.repo.Commits) - count
	}
	return firstIndex, h.repo.Commits[firstIndex : firstIndex+count]
}

func (h *repoVM) showContextMenu(x, y int) {
	menu := h.mainService.NewMenu("")

	showItems := h.GetOpenBranchMenuItems()
	menu.Add(ui.MenuItem{Text: "Show Branch", SubItems: showItems})

	hideItems := h.GetCloseBranchMenuItems()
	menu.Add(ui.MenuItem{Text: "Hide Branch", SubItems: hideItems})

	menu.Add(ui.SeparatorMenuItem)
	c := h.repo.Commits[h.firstIndex+y]
	menu.Add(ui.MenuItem{Text: "Commit Diff ...", Key: "Ctrl-D", Action: func() {
		h.ShowDiff(c.ID)
	}})
	menu.Add(ui.MenuItem{Text: "Commit ...", Key: "Ctrl-Space", Action: func() {
		h.commit()
	}})
	switchItems := h.GetSwitchBranchMenuItems()
	menu.Add(ui.MenuItem{Text: "Switch/Checkout", SubItems: switchItems})
	menu.Add(h.mainService.RecentReposMenuItem())
	menu.Add(h.mainService.MainMenuItem())
	//
	menu.Show(x+1, y+2)
}

func (h *repoVM) GetOpenBranchMenuItems() []ui.MenuItem {
	branches := h.viewModelService.GetCommitOpenBranches(h.currentIndex, h.repo)

	current, ok := h.viewModelService.CurrentBranch(h.repo)
	if ok {
		if nil == funk.Find(branches, func(b viewmodel.Branch) bool {
			return current.DisplayName == b.DisplayName
		}) {
			branches = append(branches, current)
		}
	}

	var items []ui.MenuItem
	for _, b := range branches {
		items = append(items, h.toOpenBranchMenuItem(b))
	}

	if len(items) > 0 {
		items = append(items, ui.SeparatorMenuItem)
	}

	var activeSubItems []ui.MenuItem
	for _, b := range h.viewModelService.GetActiveBranches(h.repo) {
		activeSubItems = append(activeSubItems, h.toOpenBranchMenuItem(b))
	}
	items = append(items, ui.MenuItem{Text: "Active Branches", SubItems: activeSubItems})

	var allGitSubItems []ui.MenuItem
	for _, b := range h.viewModelService.GetAllBranches(h.repo) {
		if b.IsGitBranch {
			allGitSubItems = append(allGitSubItems, h.toOpenBranchMenuItem(b))
		}
	}
	items = append(items, ui.MenuItem{Text: "All Git Branches", SubItems: allGitSubItems})

	var allSubItems []ui.MenuItem
	for _, b := range h.viewModelService.GetAllBranches(h.repo) {
		allSubItems = append(allSubItems, h.toOpenBranchMenuItem(b))
	}
	items = append(items, ui.MenuItem{Text: "All Branches", SubItems: allSubItems})

	return items

}

func (h *repoVM) GetCloseBranchMenuItems() []ui.MenuItem {
	var items []ui.MenuItem
	commitBranches := h.viewModelService.GetShownBranches(h.repo, true)
	for _, b := range commitBranches {
		items = append(items, h.toCloseBranchMenuItem(b))
	}
	return items
}

func (h *repoVM) GetSwitchBranchMenuItems() []ui.MenuItem {
	var items []ui.MenuItem
	commitBranches := h.viewModelService.GetShownBranches(h.repo, false)
	for _, b := range commitBranches {
		items = append(items, h.toSwitchBranchMenuItem(b))
	}
	return items
}

func (h *repoVM) toOpenBranchMenuItem(branch viewmodel.Branch) ui.MenuItem {
	return ui.MenuItem{Text: h.branchItemText(branch), Action: func() {
		h.viewModelService.ShowBranch(branch.Name, h.repo)
	}}
}

func (h *repoVM) toCloseBranchMenuItem(branch viewmodel.Branch) ui.MenuItem {
	return ui.MenuItem{Text: h.branchItemText(branch), Action: func() {
		h.viewModelService.HideBranch(h.repo, branch.Name)
	}}
}

func (h *repoVM) toSwitchBranchMenuItem(branch viewmodel.Branch) ui.MenuItem {
	return ui.MenuItem{Text: h.branchItemText(branch), Action: func() {
		h.viewModelService.SwitchToBranch(branch.Name)
	}}
}

func (h *repoVM) branchItemText(branch viewmodel.Branch) string {
	if branch.IsCurrent {
		return "●" + branch.DisplayName
	} else {
		return " " + branch.DisplayName
	}
}

func (h *repoVM) refresh() {
	log.Infof("refresh")
	h.viewModelService.TriggerRefreshModel()
}

func (h *repoVM) RefreshTrace(viewPage ui.ViewPage) {
	git.EnableTracing("")
	// traceBytes := utils.MustJsonMarshal(trace{
	// 	RepoPath:    h.viewModelService.RepoPath(),
	// 	ViewPage:    viewPage,
	// 	BranchNames: h.viewModelService.CurrentBranchNames(),
	// })
	// utils.MustFileWrite(filepath.Join(git.CurrentTracePath(), "repovm"), traceBytes)

	//h.viewModelService.TriggerRefreshModel()
}

func (h *repoVM) ChangeBranchColor() {
	//h.viewModelService.ChangeBranchColor(index)
}

func (h *repoVM) ToggleDetails() {
	h.isDetails = !h.isDetails
	if h.isDetails {
		log.Event("details-view-show")
	} else {
		log.Event("details-view-hide")
	}
}

func (h *repoVM) showDiff() {
	c := h.repo.Commits[h.currentIndex]
	h.ShowDiff(c.ID)
}

func (h *repoVM) saveTotalDebugState() {
	//	h.vm.RefreshTrace(h.ViewPage())
}

func (h *repoVM) commit() {
	commitView := NewCommitView(h.ui, h.viewModelService)
	commitView.Show()
}

func (h *repoVM) ShowDiff(commitID string) {
	diffView := NewDiffView(h.ui, h.viewModelService, commitID)
	diffView.Show()
	diffView.SetTop()
	diffView.SetCurrentView()
}
