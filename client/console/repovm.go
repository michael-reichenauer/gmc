package console

import (
	"context"
	"fmt"
	"strings"

	"github.com/michael-reichenauer/gmc/api"
	"github.com/michael-reichenauer/gmc/utils/cui"
	"github.com/michael-reichenauer/gmc/utils/git"
	"github.com/michael-reichenauer/gmc/utils/log"
	"github.com/samber/lo"
)

// repoPage
type repoPage struct {
	lines              []string
	total              int
	repoPath           string
	currentBranchName  string
	uncommittedChanges int
	selectedBranchName string
}

type RepoViewer interface {
	NotifyChanged()
	ShowLineAtTop(line int)
	OpenRepoMenuItems() []cui.MenuItem
	ShowSearchView()
}

type repoVM struct {
	ui                cui.UI
	repoViewer        RepoViewer
	api               api.Api
	repoLayout        *repoLayout
	isDetails         bool
	cancel            context.CancelFunc
	repo              api.Repo
	searchRepo        api.Repo
	firstIndex        int
	currentIndex      int
	onRepoUpdatedFunc func()
	searchText        string
	done              chan struct{}
	repoID            string
}

type trace struct {
	RepoPath    string
	ViewPage    cui.ViewPage
	BranchNames []string
}

func newRepoVM(ui cui.UI, repoViewer RepoViewer, api api.Api, repoID string) *repoVM {
	return &repoVM{
		ui:         ui,
		repoViewer: repoViewer,
		api:        api,
		repoID:     repoID,
		repoLayout: newRepoLayout(),
		done:       make(chan struct{}),
	}
}

func (t *repoVM) startRepoMonitor() {
	go t.monitorModelRoutine()
}

func (t *repoVM) triggerRefresh() {
	log.Event("repoview-refresh")
	progress := t.ui.ShowProgress("Trigger")
	t.startCommand(
		fmt.Sprintf("Trigger refresh repo"),
		func() error { return t.api.TriggerRefreshRepo(t.repoID, api.NilRsp) },
		func(err error) string { return fmt.Sprintf("Failed to trigger:\n%v", err) },
		func() {
			t.ui.Post(func() {
				progress.Close()
			})
		})
}

func (t *repoVM) SetSearch(text string) {
	t.startCommand(
		fmt.Sprintf("Trigger search repo"),
		func() error { return t.api.TriggerSearch(api.Search{RepoID: t.repoID, Text: text}, api.NilRsp) },
		func(err error) string { return fmt.Sprintf("Failed to trigger:\n%v", err) },
		nil)
}

func (t *repoVM) close() {
	log.Infof("Close")
	close(t.done)
	_ = t.api.CloseRepo(t.repoID, api.NilRsp)
}

func (t *repoVM) monitorModelRoutine() {
	repoChanges := make(chan api.RepoChange)
	go func() {
		for {
			var changes []api.RepoChange
			err := t.api.GetRepoChanges(t.repoID, &changes)
			if err != nil {
				close(repoChanges)
				return
			}
			select {
			case <-t.done:
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
		log.Infof("repo event")
		rc := r
		t.ui.Post(func() {
			log.Debugf("Repo change event:")
			if progress != nil {
				log.Debugf("Repo change event: closing previous progress")
				progress.Close()
				progress = nil
			}
			if rc.IsStarting {
				log.Debugf("Repo change event: repo starting event")
				progress = t.ui.ShowProgress("Loading repo")
				return
			}
			log.Debugf("Repo change event (not starting event)")

			if rc.Error != nil {
				log.Warnf("Repo change event: repo error event")
				t.ui.ShowErrorMessageBox("Error: %v", rc.Error)
				return
			}

			if rc.SearchText != "" {
				log.Infof("repo search event")
				log.Infof("commits %d", len(r.ViewRepo.Commits))
				t.repo = r.ViewRepo
				t.repoViewer.NotifyChanged()
				return
			}

			t.repo = rc.ViewRepo
			t.repoViewer.NotifyChanged()

			if t.onRepoUpdatedFunc != nil {
				f := t.onRepoUpdatedFunc
				t.onRepoUpdatedFunc = nil
				t.ui.Post(f)
			}
		})
	}
}

func (t *repoVM) GetRepoPage(viewPage cui.ViewPage) (repoPage, error) {
	var sbn string
	if viewPage.CurrentLine < len(t.repo.Commits) {
		sc := t.repo.Commits[viewPage.CurrentLine]
		sbn = t.repo.Branches[sc.BranchIndex].DisplayName
	}

	firstIndex, lines := t.getLines(viewPage, sbn)
	t.firstIndex = firstIndex
	t.currentIndex = viewPage.CurrentLine

	return repoPage{
		repoPath:           t.repo.RepoPath,
		lines:              lines,
		total:              len(t.repo.Commits),
		uncommittedChanges: t.repo.UncommittedChanges,
		currentBranchName:  t.repo.CurrentBranchName,
		selectedBranchName: sbn,
	}, nil
}

func (t *repoVM) getLines(viewPage cui.ViewPage, selectedBranchName string) (int, []string) {
	firstIndex, commits, graph := t.getPage(viewPage)
	return firstIndex, t.repoLayout.getPageLines(commits, graph, viewPage.Width, selectedBranchName, t.repo)
}

func (t *repoVM) isMoreClick(x int, y int) bool {
	moreX := t.repoLayout.getMoreIndex(t.repo)
	return x == moreX
}

func (t *repoVM) getPage(viewPage cui.ViewPage) (int, []api.Commit, []api.GraphRow) {
	firstIndex := viewPage.FirstLine
	count := viewPage.Height
	if count > len(t.repo.Commits) {
		// Requested count larger than available, return just all available commits
		count = len(t.repo.Commits)
	}

	if firstIndex+count >= len(t.repo.Commits) {
		// Requested commits past available, adjust to return available commits
		firstIndex = len(t.repo.Commits) - count
	}
	commits := t.repo.Commits[firstIndex : firstIndex+count]
	graphRows := t.repo.ConsoleGraph[firstIndex : firstIndex+count]
	return firstIndex, commits, graphRows
}

func (t *repoVM) showCommitDetails() {
	c := t.repo.Commits[t.currentIndex]

	var cd api.CommitDetailsRsp
	err := t.api.GetCommitDetails(api.CommitDetailsReq{RepoID: t.repoID, CommitID: c.ID}, &cd)
	if err != nil {
		log.Warnf("Failed: %v", err)
		return
	}
	files := strings.Join(cd.Files, "\n")
	title := fmt.Sprintf("Commit %s", c.SID)
	id := c.ID

	if c.ID == git.UncommittedID {
		title = "Uncommitted"
		id = ""
	}

	text := fmt.Sprintf(cui.Blue("Id:")+"       %s\n"+
		cui.Blue("Branches:")+" %s\n"+
		"%s\n\n"+
		cui.Blue(strings.Repeat("_", 50))+
		cui.Blue("\n%d Files:\n")+
		"%s",
		id, cd.BranchName, cd.Message, len(cd.Files), files)

	t.ui.ShowMessageBox(title, text)
}

func (t *repoVM) showCommitDialog() {
	if t.repo.Conflicts > 0 {
		t.ui.ShowErrorMessageBox("Conflicts must be resolved before committing.")
		return
	}

	commitView := NewCommitView(t.ui, t.api, t.repoID, t.repo.CurrentBranchName)
	message := t.repo.MergeMessage
	commitView.Show(message)
}

func (t *repoVM) showCreateBranchDialog() {
	branchView := NewBranchView(t.ui, t)
	branchView.Show()
}

func (t *repoVM) showCommitDiff(commitID string) {
	diffView := NewCommitDiffView(t.ui, t.api, t.repoID, commitID)
	diffView.Show()
}

func (t *repoVM) showFileDiff(path string) {
	diffView := NewFileDiffView(t.ui, t.api, t.repoID, path)
	diffView.Show()
}

func (t *repoVM) ShowSearchView() {
	t.repoViewer.ShowSearchView()
}

func (t *repoVM) showSelectedSearchCommit() {
	c := t.repo.Commits[t.currentIndex]
	t.showCommitDiff(c.ID)
}

func (t *repoVM) showSelectedCommitDiff() {
	c := t.repo.Commits[t.currentIndex]
	t.showCommitDiff(c.ID)
}

func (t *repoVM) GetCommitBranches(selectedIndex int) []api.Branch {
	c := t.repo.Commits[selectedIndex]
	if c.More == api.MoreNone {
		return nil
	}

	var branches []api.Branch
	_ = t.api.GetBranches(api.GetBranchesReq{RepoID: t.repoID, IncludeOnlyCommitBranches: c.ID}, &branches)

	return branches
}

func (t *repoVM) GetFiles(ref string) []string {
	var files []string
	_ = t.api.GetFiles(api.FilesReq{RepoID: t.repoID, Ref: ref}, &files)

	return files
}

func (t *repoVM) CurrentNotShownBranch() (api.Branch, bool) {
	var branches []api.Branch
	err := t.api.GetBranches(
		api.GetBranchesReq{RepoID: t.repoID, IncludeOnlyCurrent: true, IncludeOnlyNotShown: true},
		&branches)

	if err != nil || len(branches) == 0 {
		return api.Branch{}, false
	}

	return branches[0], true
}

func (t *repoVM) CurrentBranch() (api.Branch, bool) {
	var branches []api.Branch
	err := t.api.GetBranches(
		api.GetBranchesReq{RepoID: t.repoID, IncludeOnlyCurrent: true},
		&branches)

	if err != nil || len(branches) == 0 {
		return api.Branch{}, false
	}

	return branches[0], true
}

func (t *repoVM) GetLatestBranches(skipShown bool) []api.Branch {
	var branches []api.Branch

	_ = t.api.GetBranches(api.GetBranchesReq{
		RepoID:              t.repoID,
		IncludeOnlyNotShown: skipShown,
		SortOnLatest:        true,
	}, &branches)
	return branches
}

func (t *repoVM) GetAllBranches(skipShown bool) []api.Branch {
	var branches []api.Branch

	_ = t.api.GetBranches(api.GetBranchesReq{RepoID: t.repoID, IncludeOnlyNotShown: skipShown}, &branches)
	return branches
}

func (t *repoVM) GetShownBranches(skipMaster bool) []api.Branch {
	var branches []api.Branch
	_ = t.api.GetBranches(
		api.GetBranchesReq{RepoID: t.repoID, IncludeOnlyShown: true, SkipMaster: skipMaster},
		&branches)
	return branches
}

func (t *repoVM) GetNotShownAmbiguousBranches() []api.Branch {
	var branches []api.Branch

	_ = t.api.GetBranches(api.GetBranchesReq{RepoID: t.repoID, IncludeOnlyNotShown: true}, &branches)

	var bs []api.Branch
	for _, b := range branches {
		if b.IsAmbiguousBranch {
			bs = append(bs, b)
		}
	}
	return bs
}

func (t *repoVM) ShowBranch(name string) {
	t.startCommand(
		fmt.Sprintf("Show Branch:\n%s", name), func() error {
			return t.api.ShowBranch(api.BranchName{RepoID: t.repoID, BranchName: name}, api.NilRsp)
		},
		func(err error) string { return fmt.Sprintf("Failed to show branch:\n%s\n%s", name, err) },
		func() { t.ScrollToBranch(name) })
}

func (t *repoVM) ScrollToBranch(name string) {
	t.ui.Post(func() {
		branch, ok := lo.Find(t.repo.Branches, func(v api.Branch) bool { return v.Name == name })
		if !ok {
			return
		}

		_, i, ok := lo.FindIndexOf(t.repo.Commits, func(v api.Commit) bool { return v.ID == branch.TipID })
		if !ok {
			return
		}

		t.repoViewer.ShowLineAtTop(i)
	})
}

func (t *repoVM) SetAsParentBranch(name string) {
	_ = t.api.SetAsParentBranch(api.BranchName{RepoID: t.repoID, BranchName: name}, api.NilRsp)
}

func (t *repoVM) UnsetAsParentBranch(name string) {
	_ = t.api.UnsetAsParentBranch(api.BranchName{RepoID: t.repoID, BranchName: name}, api.NilRsp)
}

func (t *repoVM) HideBranch(name string) {
	_ = t.api.HideBranch(api.BranchName{RepoID: t.repoID, BranchName: name}, api.NilRsp)
}

func (t *repoVM) SwitchToBranch(name string, displayName string) {
	t.startCommand(
		fmt.Sprintf("Switch/checkout:\n%s", name),
		func() error {
			return t.api.Checkout(api.CheckoutReq{RepoID: t.repoID, Name: name, DisplayName: displayName}, api.NilRsp)
		},
		func(err error) string { return fmt.Sprintf("Failed to switch/checkout:\n%s\n%s", name, err) },
		nil)
}

func (t *repoVM) PushBranch(name string) {
	t.startCommand(
		fmt.Sprintf("Pushing Branch:\n%s", name),
		func() error { return t.api.PushBranch(api.BranchName{RepoID: t.repoID, BranchName: name}, api.NilRsp) },
		func(err error) string { return fmt.Sprintf("Failed to push:\n%s\n%s", name, err) },
		nil)
}

func (t *repoVM) PushCurrentBranch() {
	current, ok := t.CurrentBranch()
	if !ok || !current.HasLocalOnly {
		return
	}
	t.startCommand(
		fmt.Sprintf("Pushing current branch:\n%s", current.Name),
		func() error {
			return t.api.PushBranch(api.BranchName{RepoID: t.repoID, BranchName: current.Name}, api.NilRsp)
		},
		func(err error) string { return fmt.Sprintf("Failed to push:\n%s\n%s", current.Name, err) },
		nil)
}

func (t *repoVM) PullCurrentBranch() {
	current, ok := t.CurrentBranch()
	if !ok {
		return
	}

	t.startCommand(
		fmt.Sprintf("Pull/Update current branch:\n%s", current.Name),
		func() error { return t.api.PullCurrentBranch(t.repoID, api.NilRsp) },
		func(err error) string { return fmt.Sprintf("Failed to pull/update:\n%s\n%s", current.Name, err) },
		nil)
}

func (t *repoVM) PullBranch(name string) {
	log.Infof("Pull branch %q", name)
	t.startCommand(
		fmt.Sprintf("Pull/Update branch:\n%s", name),
		func() error { return t.api.PullBranch(api.BranchName{RepoID: t.repoID, BranchName: name}, api.NilRsp) },
		func(err error) string { return fmt.Sprintf("Failed to pull/update:\n%s\n%s", name, err) },
		nil)
}

func (t *repoVM) MergeFromBranch(name string) {
	t.startCommand(
		fmt.Sprintf("Merging to Branch:\n%s", name),
		func() error { return t.api.MergeBranch(api.BranchName{RepoID: t.repoID, BranchName: name}, api.NilRsp) },
		func(err error) string { return fmt.Sprintf("Failed to merge:\n%s\n%s", name, err) },
		nil)
}

func (t *repoVM) startCommand(
	progressText string,
	doFunc func() error,
	errorFunc func(err error) string,
	onRepoUpdatedFunc func(),
) {
	progress := t.ui.ShowProgress(progressText)
	t.onRepoUpdatedFunc = onRepoUpdatedFunc
	go func() {
		err := doFunc()
		t.ui.Post(func() {
			if err != nil {
				t.ui.ShowErrorMessageBox(errorFunc(err))
			}
			progress.Close()
		})
	}()
}

func (t *repoVM) CreateBranch(name string) {
	t.startCommand(
		fmt.Sprintf("Creating Branch:\n%s", name),
		func() error {
			err := t.api.CreateBranch(api.BranchName{RepoID: t.repoID, BranchName: name}, api.NilRsp)
			if err != nil {
				return err
			}
			err = t.api.PushBranch(api.BranchName{RepoID: t.repoID, BranchName: name}, api.NilRsp)
			if err != nil {
				return err
			}
			return err
		},
		func(err error) string { return fmt.Sprintf("Failed to create branch:\n%s\n%s", name, err) },
		func() { t.ShowBranch(name) })
}

func (t *repoVM) DeleteBranch(name string) {
	t.startCommand(
		fmt.Sprintf("Deleting Branch:\n%s", name),
		func() error {
			return t.api.DeleteBranch(api.BranchName{RepoID: t.repoID, BranchName: name}, api.NilRsp)
		},
		func(err error) string { return fmt.Sprintf("Failed to delete:\n%s\n%s", name, err) },
		nil)
}

func (t *repoVM) GetAmbiguousBranchBranchesMenuItems() []api.Branch {
	commit := t.repo.Commits[t.currentIndex]
	branch := t.repo.Branches[commit.BranchIndex]
	if !branch.IsAmbiguousBranch {
		return nil
	}

	var branches []api.Branch
	_ = t.api.GetAmbiguousBranchBranches(api.AmbiguousBranchBranchesReq{RepoID: t.repoID, CommitID: commit.ID}, &branches)

	return branches
}
