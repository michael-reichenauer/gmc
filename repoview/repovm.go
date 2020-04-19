package repoview

import (
	"context"
	"fmt"
	"github.com/michael-reichenauer/gmc/common/config"
	"github.com/michael-reichenauer/gmc/repoview/viewmodel"
	"github.com/michael-reichenauer/gmc/utils/git"
	"github.com/michael-reichenauer/gmc/utils/log"
	"github.com/michael-reichenauer/gmc/utils/ui"
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
	repoViewer       *RepoView
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
	ui *ui.UI,
	repoViewer *RepoView,
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
	ctx, cancel := context.WithCancel(context.Background())
	h.cancel = cancel
	go h.monitorModelRoutine(ctx)
}

func (h *repoVM) close() {
	h.cancel()
}

func (h *repoVM) monitorModelRoutine(ctx context.Context) {
	h.viewModelService.StartMonitor(ctx)
	var progress *ui.Progress
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

func (h *repoVM) refresh() {
	log.Event("repoview-refresh")
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

func (h *repoVM) commit() {
	commitView := NewCommitView(h.ui, h.viewModelService, h)
	commitView.Show()
}

func (h *repoVM) showCommitDiff(commitID string) {
	diffView := NewDiffView(h.ui, h.viewModelService, commitID)
	diffView.Show()
	diffView.SetTop()
	diffView.SetCurrentView()
}

func (h *repoVM) showSelectedCommitDiff() {
	c := h.repo.Commits[h.currentIndex]
	h.showCommitDiff(c.ID)
}

func (h *repoVM) GetCommitOpenBranches() []viewmodel.Branch {
	c := h.repo.Commits[h.currentIndex]
	if c.More == viewmodel.MoreNone {
		return nil
	}

	return h.viewModelService.GetCommitOpenBranches(c.ID, h.repo)
}

func (h *repoVM) CurrentNotShownBranch() (viewmodel.Branch, bool) {
	current, ok := h.viewModelService.CurrentNotShownBranch(h.repo)

	return current, ok
}

func (h *repoVM) CurrentBranch() (viewmodel.Branch, bool) {
	current, ok := h.viewModelService.CurrentBranch(h.repo)
	return current, ok
}

func (h *repoVM) GetActiveBranches() []viewmodel.Branch {
	return h.viewModelService.GetActiveBranches(h.repo)
}

func (h *repoVM) GetAllBranches() []viewmodel.Branch {
	return h.viewModelService.GetAllBranches(h.repo)
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

func (h *repoVM) SwitchToBranch(name string) {
	h.startCommand(
		fmt.Sprintf("Switch/checkout\n%s", name),
		func() error { return h.viewModelService.SwitchToBranch(name) },
		func(err error) string { return fmt.Sprintf("Failed to switch/checkout:\n%s\n%s", name, err) })
}

func (h *repoVM) PushBranch(name string) {
	h.startCommand(
		fmt.Sprintf("Pushing Branch\n%s", name),
		func() error { return h.viewModelService.PushBranch(name) },
		func(err error) string { return fmt.Sprintf("Failed to push:\n%s\n%s", name, err) })
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
