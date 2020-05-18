package console

import (
	"context"
	"fmt"
	"github.com/michael-reichenauer/gmc/common/config"
	"github.com/michael-reichenauer/gmc/utils/cui"
	"github.com/michael-reichenauer/gmc/utils/log"
	"github.com/michael-reichenauer/gmc/viewrepo"
)

type repoPage struct {
	lines              []string
	total              int
	repoPath           string
	currentBranchName  string
	uncommittedChanges int
	selectedBranchName string
}

type repoVM struct {
	ui                cui.UI
	repoViewer        cui.Notifier
	mainService       mainService
	viewModelService  *viewrepo.Service
	repoLayout        *repoLayout
	isDetails         bool
	workingFolder     string
	cancel            context.CancelFunc
	repo              viewrepo.ViewRepo
	searchRepo        viewrepo.ViewRepo
	firstIndex        int
	currentIndex      int
	onRepoUpdatedFunc func()
	searchText        string
}

type trace struct {
	RepoPath    string
	ViewPage    cui.ViewPage
	BranchNames []string
}

func newRepoVM(
	ui cui.UI,
	repoViewer cui.Notifier,
	mainService mainService,
	configService *config.Service,
	workingFolder string) *repoVM {
	viewModelService := viewrepo.NewService(configService, workingFolder)
	return &repoVM{
		ui:               ui,
		repoViewer:       repoViewer,
		mainService:      mainService,
		viewModelService: viewModelService,
		repoLayout:       newRepoLayout(viewModelService),
		workingFolder:    workingFolder,
	}
}

func (h *repoVM) startRepoMonitor() {
	ctx, cancel := context.WithCancel(context.Background())
	h.cancel = cancel
	go h.monitorModelRoutine(ctx)
}

func (h *repoVM) triggerRefresh() {
	log.Event("repoview-refresh")
	h.viewModelService.TriggerRefreshModel()
}

func (h *repoVM) SetSearch(text string) {
	h.viewModelService.TriggerSearch(text)
}

func (h *repoVM) close() {
	h.cancel()
}

func (h *repoVM) monitorModelRoutine(ctx context.Context) {
	h.viewModelService.StartMonitor(ctx)
	var progress cui.Progress
	for r := range h.viewModelService.RepoChanges {
		log.Infof("Detected model change")
		rc := r
		h.ui.PostOnUIThread(func() {
			if rc.IsStarting {
				progress = h.ui.ShowProgress("Loading repo:\n%s", h.workingFolder)
				return
			}

			if progress != nil {
				progress.Close()
				progress = nil
			}

			if rc.Error != nil {
				h.ui.ShowErrorMessageBox("Error: %v", rc.Error)
				return
			}

			if rc.SearchText != "" {
				log.Infof("commits %d", len(r.ViewRepo.Commits))
				h.repo = r.ViewRepo
				h.repoViewer.NotifyChanged()
				return
			}

			h.repo = rc.ViewRepo
			h.repoViewer.NotifyChanged()

			if h.onRepoUpdatedFunc != nil {
				f := h.onRepoUpdatedFunc
				h.onRepoUpdatedFunc = nil
				h.ui.PostOnUIThread(f)
			}
		})
	}
}

func (h *repoVM) GetRepoPage(viewPage cui.ViewPage) (repoPage, error) {
	firstIndex, lines := h.getLines(viewPage)
	h.firstIndex = firstIndex
	h.currentIndex = viewPage.CurrentLine

	var sbn string
	if viewPage.CurrentLine < len(h.repo.Commits) {
		sbn = h.repo.Commits[viewPage.CurrentLine].Branch.Name
	}
	return repoPage{
		repoPath:           h.repo.RepoPath,
		lines:              lines,
		total:              len(h.repo.Commits),
		uncommittedChanges: h.repo.UncommittedChanges,
		currentBranchName:  h.repo.CurrentBranchName,
		selectedBranchName: sbn,
	}, nil
}

func (h *repoVM) getLines(viewPage cui.ViewPage) (int, []string) {
	firstIndex, commits := h.getCommits(viewPage)
	return firstIndex, h.repoLayout.getPageLines(commits, viewPage.Width, "", h.repo)
}

func (h *repoVM) isMoreClick(x int, y int) bool {
	moreX := h.repoLayout.getMoreIndex(h.repo)
	return x == moreX
}

func (h *repoVM) getCommits(viewPage cui.ViewPage) (int, []viewrepo.Commit) {
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

func (h *repoVM) showCommitDialog() {
	if h.repo.Conflicts > 0 {
		h.ui.ShowErrorMessageBox("Conflicts must be resolved before committing.")
		return
	}
	commitView := NewCommitView(h.ui, h.viewModelService)
	message := h.repo.MergeMessage
	commitView.Show(message)
}

func (h *repoVM) showCreateBranchDialog() {
	branchView := NewBranchView(h.ui, h)
	branchView.Show()
}

func (h *repoVM) showCommitDiff(commitID string) {
	diffView := NewDiffView(h.ui, h.viewModelService, commitID)
	diffView.Show()
}

func (h *repoVM) showSelectedCommitDiff() {
	c := h.repo.Commits[h.currentIndex]
	h.showCommitDiff(c.ID)
}

func (h *repoVM) GetCommitOpenInBranches(selectedIndex int) []viewrepo.Branch {
	c := h.repo.Commits[selectedIndex]
	if c.More == viewrepo.MoreNone {
		return nil
	}

	return h.viewModelService.GetCommitOpenInBranches(c.ID, h.repo)
}

func (h *repoVM) GetCommitOpenOutBranches(selectedIndex int) []viewrepo.Branch {
	c := h.repo.Commits[selectedIndex]
	if c.More == viewrepo.MoreNone {
		return nil
	}

	return h.viewModelService.GetCommitOpenOutBranches(c.ID, h.repo)
}

func (h *repoVM) CurrentNotShownBranch() (viewrepo.Branch, bool) {
	current, ok := h.viewModelService.CurrentNotShownBranch(h.repo)

	return current, ok
}

func (h *repoVM) CurrentBranch() (viewrepo.Branch, bool) {
	current, ok := h.viewModelService.CurrentBranch(h.repo)
	return current, ok
}

func (h *repoVM) GetLatestBranches(skipShown bool) []viewrepo.Branch {
	return h.viewModelService.GetLatestBranches(h.repo, skipShown)
}

func (h *repoVM) GetAllBranches(skipShown bool) []viewrepo.Branch {
	return h.viewModelService.GetAllBranches(h.repo, skipShown)
}

func (h *repoVM) GetShownBranches(skipMaster bool) []viewrepo.Branch {
	return h.viewModelService.GetShownBranches(h.repo, skipMaster)
}

func (h *repoVM) ShowBranch(name string) {
	h.viewModelService.ShowBranch(name, h.repo)
}

func (h *repoVM) HideBranch(name string) {
	h.viewModelService.HideBranch(h.repo, name)
}

func (h *repoVM) SwitchToBranch(name string, displayName string) {
	h.startCommand(
		fmt.Sprintf("Switch/checkout:\n%s", name),
		func() error { return h.viewModelService.SwitchToBranch(name, displayName, h.repo) },
		func(err error) string { return fmt.Sprintf("Failed to switch/checkout:\n%s\n%s", name, err) },
		nil)
}

func (h *repoVM) PushBranch(name string) {
	h.startCommand(
		fmt.Sprintf("Pushing Branch:\n%s", name),
		func() error { return h.viewModelService.PushBranch(name) },
		func(err error) string { return fmt.Sprintf("Failed to push:\n%s\n%s", name, err) },
		nil)
}

func (h *repoVM) PushCurrentBranch() {
	current, ok := h.CurrentBranch()
	if !ok || !current.HasLocalOnly {
		return
	}
	h.startCommand(
		fmt.Sprintf("Pushing current branch:\n%s", current.Name),
		func() error { return h.viewModelService.PushBranch(current.Name) },
		func(err error) string { return fmt.Sprintf("Failed to push:\n%s\n%s", current.Name, err) },
		nil)
}

func (h *repoVM) PullCurrentBranch() {
	current, ok := h.CurrentBranch()
	if !ok || !current.HasLocalOnly {
		return
	}
	h.startCommand(
		fmt.Sprintf("Pull/Update current branch:\n%s", current.Name),
		func() error { return h.viewModelService.PullBranch() },
		func(err error) string { return fmt.Sprintf("Failed to pull/update:\n%s\n%s", current.Name, err) },
		nil)
}

func (h *repoVM) MergeFromBranch(name string) {
	h.startCommand(
		fmt.Sprintf("Merging to Branch:\n%s", name),
		func() error { return h.viewModelService.MergeBranch(name) },
		func(err error) string { return fmt.Sprintf("Failed to merge:\n%s\n%s", name, err) },
		nil)
}

func (h *repoVM) startCommand(
	progressText string,
	doFunc func() error,
	errorFunc func(err error) string,
	onRepoUpdatedFunc func()) {
	progress := h.ui.ShowProgress(progressText)
	h.onRepoUpdatedFunc = onRepoUpdatedFunc
	go func() {
		err := doFunc()
		h.ui.PostOnUIThread(func() {
			if err != nil {
				h.ui.ShowErrorMessageBox(errorFunc(err))
			}
			progress.Close()
		})
	}()
}

func (h *repoVM) CreateBranch(name string) {
	h.startCommand(
		fmt.Sprintf("Creating Branch:\n%s", name),
		func() error {
			err := h.viewModelService.CreateBranch(name)
			if err != nil {
				return err
			}
			err = h.viewModelService.PushBranch(name)
			if err != nil {
				return err
			}
			return err
		},
		func(err error) string { return fmt.Sprintf("Failed to create branch:\n%s\n%s", name, err) },
		func() { h.viewModelService.ShowBranch(name, h.repo) })
}

func (h *repoVM) DeleteBranch(name string) {
	h.startCommand(
		fmt.Sprintf("Deleting Branch:\n%s", name),
		func() error { return h.viewModelService.DeleteBranch(name, h.repo) },
		func(err error) string { return fmt.Sprintf("Failed to delete:\n%s\n%s", name, err) },
		nil)
}
