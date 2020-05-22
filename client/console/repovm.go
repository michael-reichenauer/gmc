package console

import (
	"context"
	"fmt"
	"github.com/michael-reichenauer/gmc/api"
	"github.com/michael-reichenauer/gmc/server/viewrepo"
	"github.com/michael-reichenauer/gmc/utils/cui"
	"github.com/michael-reichenauer/gmc/utils/log"
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
	viewRepo          api.Repo
	repoLayout        *repoLayout
	isDetails         bool
	cancel            context.CancelFunc
	repo              viewrepo.Repo
	searchRepo        viewrepo.Repo
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
	viewRepo api.Repo,
) *repoVM {
	return &repoVM{
		ui:         ui,
		repoViewer: repoViewer,
		viewRepo:   viewRepo,
		repoLayout: newRepoLayout(),
	}
}

func (h *repoVM) startRepoMonitor() {
	go h.monitorModelRoutine()
}

func (h *repoVM) triggerRefresh() {
	log.Event("repoview-refresh")
	h.viewRepo.TriggerRefreshModel()
}

func (h *repoVM) SetSearch(text string) {
	h.viewRepo.TriggerSearch(text)
}

func (h *repoVM) close() {
	h.viewRepo.Close()
}

func (h *repoVM) monitorModelRoutine() {
	h.viewRepo.StartMonitor()
	var progress cui.Progress
	for r := range h.viewRepo.RepoChanges() {
		log.Infof("Detected model change")
		rc := r
		h.ui.PostOnUIThread(func() {
			if rc.IsStarting {
				progress = h.ui.ShowProgress("Loading repo")
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
	commitView := NewCommitView(h.ui, h.viewRepo)
	message := h.repo.MergeMessage
	commitView.Show(message)
}

func (h *repoVM) showCreateBranchDialog() {
	branchView := NewBranchView(h.ui, h)
	branchView.Show()
}

func (h *repoVM) showCommitDiff(commitID string) {
	diffView := NewDiffView(h.ui, h.viewRepo, commitID)
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

	return h.viewRepo.GetCommitOpenInBranches(c.ID)
}

func (h *repoVM) GetCommitOpenOutBranches(selectedIndex int) []viewrepo.Branch {
	c := h.repo.Commits[selectedIndex]
	if c.More == viewrepo.MoreNone {
		return nil
	}

	return h.viewRepo.GetCommitOpenOutBranches(c.ID)
}

func (h *repoVM) CurrentNotShownBranch() (viewrepo.Branch, bool) {
	current, ok := h.viewRepo.CurrentNotShownBranch()

	return current, ok
}

func (h *repoVM) CurrentBranch() (viewrepo.Branch, bool) {
	current, ok := h.viewRepo.CurrentBranch()
	return current, ok
}

func (h *repoVM) GetLatestBranches(skipShown bool) []viewrepo.Branch {
	return h.viewRepo.GetLatestBranches(skipShown)
}

func (h *repoVM) GetAllBranches(skipShown bool) []viewrepo.Branch {
	return h.viewRepo.GetAllBranches(skipShown)
}

func (h *repoVM) GetShownBranches(skipMaster bool) []viewrepo.Branch {
	return h.viewRepo.GetShownBranches(skipMaster)
}

func (h *repoVM) ShowBranch(name string) {
	h.viewRepo.ShowBranch(name)
}

func (h *repoVM) HideBranch(name string) {
	h.viewRepo.HideBranch(name)
}

func (h *repoVM) SwitchToBranch(name string, displayName string) {
	h.startCommand(
		fmt.Sprintf("Switch/checkout:\n%s", name),
		func() error { return h.viewRepo.SwitchToBranch(name, displayName) },
		func(err error) string { return fmt.Sprintf("Failed to switch/checkout:\n%s\n%s", name, err) },
		nil)
}

func (h *repoVM) PushBranch(name string) {
	h.startCommand(
		fmt.Sprintf("Pushing Branch:\n%s", name),
		func() error { return h.viewRepo.PushBranch(name) },
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
		func() error { return h.viewRepo.PushBranch(current.Name) },
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
		func() error { return h.viewRepo.PullBranch() },
		func(err error) string { return fmt.Sprintf("Failed to pull/update:\n%s\n%s", current.Name, err) },
		nil)
}

func (h *repoVM) MergeFromBranch(name string) {
	h.startCommand(
		fmt.Sprintf("Merging to Branch:\n%s", name),
		func() error { return h.viewRepo.MergeBranch(name) },
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
			err := h.viewRepo.CreateBranch(name)
			if err != nil {
				return err
			}
			err = h.viewRepo.PushBranch(name)
			if err != nil {
				return err
			}
			return err
		},
		func(err error) string { return fmt.Sprintf("Failed to create branch:\n%s\n%s", name, err) },
		func() { h.viewRepo.ShowBranch(name) })
}

func (h *repoVM) DeleteBranch(name string) {
	h.startCommand(
		fmt.Sprintf("Deleting Branch:\n%s", name),
		func() error { return h.viewRepo.DeleteBranch(name) },
		func(err error) string { return fmt.Sprintf("Failed to delete:\n%s\n%s", name, err) },
		nil)
}
