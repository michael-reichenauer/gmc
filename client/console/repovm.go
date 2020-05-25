package console

import (
	"context"
	"fmt"
	"github.com/michael-reichenauer/gmc/api"
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
	api               api.Api
	repoLayout        *repoLayout
	isDetails         bool
	cancel            context.CancelFunc
	repo              api.ViewRepo
	searchRepo        api.ViewRepo
	firstIndex        int
	currentIndex      int
	onRepoUpdatedFunc func()
	searchText        string
	done              chan struct{}
}

type trace struct {
	RepoPath    string
	ViewPage    cui.ViewPage
	BranchNames []string
}

func newRepoVM(
	ui cui.UI,
	repoViewer cui.Notifier,
	api api.Api,
) *repoVM {
	return &repoVM{
		ui:         ui,
		repoViewer: repoViewer,
		api:        api,
		repoLayout: newRepoLayout(),
		done:       make(chan struct{}),
	}
}

func (h *repoVM) startRepoMonitor() {
	go h.monitorModelRoutine()
}

func (h *repoVM) triggerRefresh() {
	log.Event("repoview-refresh")
	_ = h.api.TriggerRefreshModel(api.Nil, api.Nil)
}

func (h *repoVM) SetSearch(text string) {
	_ = h.api.TriggerSearch(text, api.Nil)
}

func (h *repoVM) close() {
	log.Infof("Close")
	close(h.done)
	_ = h.api.CloseRepo(api.Nil, api.Nil)
}

func (h *repoVM) monitorModelRoutine() {
	repoChanges := make(chan api.RepoChange)
	go func() {
		for {
			var changes []api.RepoChange
			err := h.api.GetChanges(api.Nil, &changes)
			if err != nil {
				close(repoChanges)
				return
			}
			select {
			case <-h.done:
				close(repoChanges)
				return
			default:
			}

			for _, c := range changes {
				repoChanges <- c
			}
		}
	}()

	var progress cui.Progress
	for r := range repoChanges {
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
	firstIndex, commits, graph := h.getPage(viewPage)
	return firstIndex, h.repoLayout.getPageLines(commits, graph, viewPage.Width, "", h.repo)
}

func (h *repoVM) isMoreClick(x int, y int) bool {
	moreX := h.repoLayout.getMoreIndex(h.repo)
	return x == moreX
}

func (h *repoVM) getPage(viewPage cui.ViewPage) (int, []api.Commit, []api.GraphRow) {
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
	commits := h.repo.Commits[firstIndex : firstIndex+count]
	graphRows := h.repo.ConsoleGraph[firstIndex : firstIndex+count]
	return firstIndex, commits, graphRows
}

func (h *repoVM) showCommitDialog() {
	if h.repo.Conflicts > 0 {
		h.ui.ShowErrorMessageBox("Conflicts must be resolved before committing.")
		return
	}
	commitView := NewCommitView(h.ui, h.api)
	message := h.repo.MergeMessage
	commitView.Show(message)
}

func (h *repoVM) showCreateBranchDialog() {
	branchView := NewBranchView(h.ui, h)
	branchView.Show()
}

func (h *repoVM) showCommitDiff(commitID string) {
	diffView := NewDiffView(h.ui, h.api, commitID)
	diffView.Show()
}

func (h *repoVM) showSelectedCommitDiff() {
	c := h.repo.Commits[h.currentIndex]
	h.showCommitDiff(c.ID)
}

func (h *repoVM) GetCommitOpenInBranches(selectedIndex int) []api.Branch {
	c := h.repo.Commits[selectedIndex]
	if c.More == api.MoreNone {
		return nil
	}

	var branches []api.Branch
	_ = h.api.GetCommitOpenInBranches(c.ID, &branches)
	return branches
}

func (h *repoVM) GetCommitOpenOutBranches(selectedIndex int) []api.Branch {
	c := h.repo.Commits[selectedIndex]
	if c.More == api.MoreNone {
		return nil
	}
	var branches []api.Branch
	_ = h.api.GetCommitOpenOutBranches(c.ID, &branches)
	return branches
}

func (h *repoVM) CurrentNotShownBranch() (api.Branch, bool) {
	var current api.Branch
	err := h.api.GetCurrentNotShownBranch(api.Nil, &current)

	return current, err == nil
}

func (h *repoVM) CurrentBranch() (api.Branch, bool) {
	var current api.Branch
	err := h.api.GetCurrentBranch(api.Nil, &current)
	return current, err == nil
}

func (h *repoVM) GetLatestBranches(skipShown bool) []api.Branch {
	var branches []api.Branch
	_ = h.api.GetLatestBranches(skipShown, &branches)
	return branches
}

func (h *repoVM) GetAllBranches(skipShown bool) []api.Branch {
	var branches []api.Branch
	_ = h.api.GetAllBranches(skipShown, &branches)
	return branches
}

func (h *repoVM) GetShownBranches(skipMaster bool) []api.Branch {
	var branches []api.Branch
	_ = h.api.GetShownBranches(skipMaster, &branches)
	return branches
}

func (h *repoVM) ShowBranch(name string) {
	_ = h.api.ShowBranch(name, api.Nil)
}

func (h *repoVM) HideBranch(name string) {
	_ = h.api.HideBranch(name, api.Nil)
}

func (h *repoVM) SwitchToBranch(name string, displayName string) {
	h.startCommand(
		fmt.Sprintf("Switch/checkout:\n%s", name),
		func() error { return h.api.SwitchToBranch(api.SwitchArgs{Name: name, DisplayName: displayName}, nil) },
		func(err error) string { return fmt.Sprintf("Failed to switch/checkout:\n%s\n%s", name, err) },
		nil)
}

func (h *repoVM) PushBranch(name string) {
	h.startCommand(
		fmt.Sprintf("Pushing Branch:\n%s", name),
		func() error { return h.api.PushBranch(name, api.Nil) },
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
		func() error { return h.api.PushBranch(current.Name, api.Nil) },
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
		func() error { return h.api.PullBranch(nil, api.Nil) },
		func(err error) string { return fmt.Sprintf("Failed to pull/update:\n%s\n%s", current.Name, err) },
		nil)
}

func (h *repoVM) MergeFromBranch(name string) {
	h.startCommand(
		fmt.Sprintf("Merging to Branch:\n%s", name),
		func() error { return h.api.MergeBranch(name, api.Nil) },
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
			err := h.api.CreateBranch(name, api.Nil)
			if err != nil {
				return err
			}
			err = h.api.PushBranch(name, api.Nil)
			if err != nil {
				return err
			}
			return err
		},
		func(err error) string { return fmt.Sprintf("Failed to create branch:\n%s\n%s", name, err) },
		func() { h.api.ShowBranch(name, api.Nil) })
}

func (h *repoVM) DeleteBranch(name string) {
	h.startCommand(
		fmt.Sprintf("Deleting Branch:\n%s", name),
		func() error { return h.api.DeleteBranch(name, api.Nil) },
		func(err error) string { return fmt.Sprintf("Failed to delete:\n%s\n%s", name, err) },
		nil)
}
