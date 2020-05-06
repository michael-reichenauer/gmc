package repoview

import (
	"context"
	"fmt"
	"github.com/michael-reichenauer/gmc/common/config"
	"github.com/michael-reichenauer/gmc/repoview/viewmodel"
	"github.com/michael-reichenauer/gmc/utils/log"
	"github.com/michael-reichenauer/gmc/utils/ui"
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
	ui               ui.UI
	repoViewer       ui.Notifier
	mainService      mainService
	viewModelService *viewmodel.Service
	repoLayout       *repoLayout
	isDetails        bool
	workingFolder    string
	cancel           context.CancelFunc
	repo             viewmodel.ViewRepo
	firstIndex       int
	currentIndex     int
}

type trace struct {
	RepoPath    string
	ViewPage    ui.ViewPage
	BranchNames []string
}

func newRepoVM(
	ui ui.UI,
	repoViewer ui.Notifier,
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

func (h *repoVM) startRepoMonitor() {
	ctx, cancel := context.WithCancel(context.Background())
	h.cancel = cancel
	go h.monitorModelRoutine(ctx)
}

func (h *repoVM) triggerRefresh() {
	log.Event("repoview-refresh")
	h.viewModelService.TriggerRefreshModel()
}

func (h *repoVM) close() {
	h.cancel()
}

func (h *repoVM) monitorModelRoutine(ctx context.Context) {
	h.viewModelService.StartMonitor(ctx)
	var progress ui.Progress
	for rc := range h.viewModelService.RepoChanges {
		log.Infof("Detected model change")
		h.ui.PostOnUIThread(func() {
			if rc.IsStarting {
				log.Infof("Show progress ...")
				progress = h.ui.ShowProgress(fmt.Sprintf("Loading repo:\n%s", h.workingFolder))
				return
			}

			if progress != nil {
				log.Infof("Close progress")
				progress.Close()
				progress = nil
			}
			h.repo = rc.ViewRepo
			h.repoViewer.NotifyChanged()
		})
	}
}

func (h *repoVM) GetRepoPage(viewPage ui.ViewPage) (repoPage, error) {
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

func (h *repoVM) getLines(viewPage ui.ViewPage) (int, []string) {
	firstIndex, commits := h.getCommits(viewPage)
	return firstIndex, h.repoLayout.getPageLines(commits, viewPage.Width, "", h.repo)
}

func (h *repoVM) isMoreClick(x int, y int) bool {
	moreX := h.repoLayout.getMoreIndex(h.repo)
	return x == moreX
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

func (h *repoVM) RefreshTrace(viewPage ui.ViewPage) {
	//git.EnableTracing("")
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
	// h.isDetails = !h.isDetails
	// if h.isDetails {
	// 	log.Event("details-view-show")
	// } else {
	// 	log.Event("details-view-hide")
	// }
}

func (h *repoVM) saveTotalDebugState() {
	//	h.vm.RefreshTrace(h.ViewPage())
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

func (h *repoVM) GetCommitOpenInBranches(selectedIndex int) []viewmodel.Branch {
	c := h.repo.Commits[selectedIndex]
	if c.More == viewmodel.MoreNone {
		return nil
	}

	return h.viewModelService.GetCommitOpenInBranches(c.ID, h.repo)
}

func (h *repoVM) GetCommitOpenOutBranches(selectedIndex int) []viewmodel.Branch {
	c := h.repo.Commits[selectedIndex]
	if c.More == viewmodel.MoreNone {
		return nil
	}

	return h.viewModelService.GetCommitOpenOutBranches(c.ID, h.repo)
}

func (h *repoVM) CurrentNotShownBranch() (viewmodel.Branch, bool) {
	current, ok := h.viewModelService.CurrentNotShownBranch(h.repo)

	return current, ok
}

func (h *repoVM) CurrentBranch() (viewmodel.Branch, bool) {
	current, ok := h.viewModelService.CurrentBranch(h.repo)
	return current, ok
}

func (h *repoVM) GetLatestBranches(skipShown bool) []viewmodel.Branch {
	return h.viewModelService.GetLatestBranches(h.repo, skipShown)
}

func (h *repoVM) GetAllBranches(skipShown bool) []viewmodel.Branch {
	return h.viewModelService.GetAllBranches(h.repo, skipShown)
}

func (h *repoVM) GetShownBranches(skipMaster bool) []viewmodel.Branch {
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
		func(err error) string { return fmt.Sprintf("Failed to switch/checkout:\n%s\n%s", name, err) })
}

func (h *repoVM) PushBranch(name string) {
	h.startCommand(
		fmt.Sprintf("Pushing Branch:\n%s", name),
		func() error { return h.viewModelService.PushBranch(name) },
		func(err error) string { return fmt.Sprintf("Failed to push:\n%s\n%s", name, err) })
}

func (h *repoVM) PushCurrentBranch() {
	current, ok := h.CurrentBranch()
	if !ok || !current.HasLocalOnly {
		return
	}
	h.startCommand(
		fmt.Sprintf("Pushing current branch:\n%s", current.Name),
		func() error { return h.viewModelService.PushBranch(current.Name) },
		func(err error) string { return fmt.Sprintf("Failed to push:\n%s\n%s", current.Name, err) })
}

func (h *repoVM) PullCurrentBranch() {
	current, ok := h.CurrentBranch()
	if !ok || !current.HasLocalOnly {
		return
	}
	h.startCommand(
		fmt.Sprintf("Pull/Update current branch:\n%s", current.Name),
		func() error { return h.viewModelService.PullBranch() },
		func(err error) string { return fmt.Sprintf("Failed to pull/update:\n%s\n%s", current.Name, err) })
}

func (h *repoVM) MergeFromBranch(name string) {
	h.startCommand(
		fmt.Sprintf("Merging to Branch:\n%s", name),
		func() error { return h.viewModelService.MergeBranch(name) },
		func(err error) string { return fmt.Sprintf("Failed to merge:\n%s\n%s", name, err) })
}

func (h *repoVM) startCommand(prsText string, doFunc func() error, errorFunc func(err error) string) {
	progress := h.ui.ShowProgress(prsText)
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
		func(err error) string { return fmt.Sprintf("Failed to create branch:\n%s\n%s", name, err) })
}

func (h *repoVM) DeleteBranch(name string) {
	h.startCommand(
		fmt.Sprintf("Deleting Branch:\n%s", name),
		func() error { return h.viewModelService.DeleteBranch(name, h.repo) },
		func(err error) string { return fmt.Sprintf("Failed to delete:\n%s\n%s", name, err) })
}
