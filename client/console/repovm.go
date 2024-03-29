package console

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/michael-reichenauer/gmc/api"
	"github.com/michael-reichenauer/gmc/common/config"
	"github.com/michael-reichenauer/gmc/utils/async"
	"github.com/michael-reichenauer/gmc/utils/cui"
	"github.com/michael-reichenauer/gmc/utils/git"
	"github.com/michael-reichenauer/gmc/utils/linq"
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
	ShowRepo(path string)
	ShowSearchView()
	ShowCommitDetails()
}

type repoVM struct {
	ui                cui.UI
	repoViewer        RepoViewer
	configService     *config.Service
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

func newRepoVM(ui cui.UI, repoViewer RepoViewer, configService *config.Service, api api.Api, repoID string) *repoVM {
	return &repoVM{
		ui:            ui,
		repoViewer:    repoViewer,
		api:           api,
		repoID:        repoID,
		repoLayout:    newRepoLayout(),
		done:          make(chan struct{}),
		configService: configService,
	}
}

func (t *repoVM) startRepoMonitor() {
	go t.monitorModelRoutine()
}

func (t *repoVM) triggerRefresh() {
	_ = t.api.TriggerRefreshRepo(t.repoID)
}

func (t *repoVM) SetSearch(text string) {
	t.startCommand(
		"Trigger search repo",
		func() error { return t.api.TriggerSearch(api.Search{RepoID: t.repoID, Text: text}) },
		func(err error) string { return fmt.Sprintf("Failed to trigger:\n%v", err) },
		nil)
}

func (t *repoVM) close() {
	log.Infof("Close")
	close(t.done)
	_ = t.api.CloseRepo(t.repoID)
}

func (t *repoVM) monitorModelRoutine() {
	repoChanges := make(chan api.RepoChange)
	go func() {
		for {
			changes, err := t.api.GetRepoChanges(t.repoID)
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
	if viewPage.CurrentLine >= 0 && viewPage.CurrentLine < len(t.repo.Commits) {
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

func (t *repoVM) isGraphClick(x int, y int) bool {
	sx := t.repoLayout.getSubjectXCoordinate(t.repo)
	return x < sx
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

func (t *repoVM) showCommitDialog() {
	current, ok := t.CurrentBranch()
	if !ok || current.TipID != git.UncommittedID {
		return
	}

	if t.repo.Conflicts > 0 {
		t.ui.ShowErrorMessageBox("Conflicts must be resolved before committing.")
		return
	}

	commitView := NewCommitView(t.ui, t.api, t.repoID, t.repo.CurrentBranchName, t.repo.UncommittedChanges)
	message := t.repo.MergeMessage
	commitView.Show(message)
}

func (t *repoVM) showCreateBranchDialog() {
	branchView := newBranchDlg(t.ui, t.CreateBranch)
	branchView.Show()
}

func (t *repoVM) showCloneDialog() {
	baseBath := ""
	paths := t.configService.GetState().RecentParentFolders
	if len(paths) > 0 {
		baseBath = paths[0] + string(os.PathSeparator)
	}

	cloneView := newCloneDlg(t.ui, baseBath, t.Clone)
	cloneView.Show()
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

	branches, _ := t.api.GetBranches(api.GetBranchesReq{RepoID: t.repoID, IncludeOnlyCommitBranches: c.ID})

	return branches
}

func (t *repoVM) GetFiles(ref string) []string {
	files, _ := t.api.GetFiles(api.FilesReq{RepoID: t.repoID, Ref: ref})
	return files
}

func (t *repoVM) CurrentNotShownBranch() (api.Branch, bool) {
	branches, err := t.api.GetBranches(
		api.GetBranchesReq{RepoID: t.repoID, IncludeOnlyCurrent: true, IncludeOnlyNotShown: true})
	if err != nil || len(branches) == 0 {
		return api.Branch{}, false
	}

	return branches[0], true
}

func (t *repoVM) CurrentBranch() (api.Branch, bool) {
	branches, err := t.api.GetBranches(
		api.GetBranchesReq{RepoID: t.repoID, IncludeOnlyCurrent: true})

	if err != nil || len(branches) == 0 {
		return api.Branch{}, false
	}

	return branches[0], true
}

func (t *repoVM) GetRecentBranches() []api.Branch {
	branches, _ := t.api.GetBranches(api.GetBranchesReq{
		RepoID:              t.repoID,
		IncludeOnlyNotShown: false,
		SortOnLatest:        true,
	})
	if len(branches) > 15 {
		branches = branches[:15]
	}
	return branches
}

func (t *repoVM) GetAllBranches() []api.Branch {
	branches, _ := t.api.GetBranches(api.GetBranchesReq{RepoID: t.repoID, IncludeOnlyNotShown: false})
	return branches
}

func (t *repoVM) GetUncommittedFiles() []string {
	diff, err := t.api.GetCommitDiff(api.CommitDiffInfoReq{RepoID: t.repoID, CommitID: git.UncommittedID})
	if err != nil {
		return []string{}
	}

	return linq.Map(diff.FileDiffs, func(v api.FileDiff) string { return v.PathAfter })
}

func (t *repoVM) UndoAllUncommittedChanges() {
	async.RunE(func() error { return t.api.UndoAllUncommittedChanges(t.repoID) }).
		Catch(func(err error) { t.ui.ShowErrorMessageBox("Failed to undo all changes:\n%s", err) })
}

func (t *repoVM) UncommitLastCommit() {
	async.RunE(func() error { return t.api.UncommitLastCommit(t.repoID) }).
		Catch(func(err error) { t.ui.ShowErrorMessageBox("Failed to uncommit:\n%s", err) })
}

func (t *repoVM) UndoCommit(id string) {
	async.RunE(func() error { return t.api.UndoCommit(t.repoID, id) }).
		Catch(func(err error) { t.ui.ShowErrorMessageBox("Failed to undo commit:\n%s:\n%s", id, err) })
}

func (t *repoVM) UndoUncommittedFileChanges(path string) {
	async.RunE(func() error { return t.api.UndoUncommittedFileChanges(t.repoID, path) }).
		Catch(func(err error) { t.ui.ShowErrorMessageBox("Failed to undo file:\n%s:\n%s", path, err) })
}

func (t *repoVM) CleanWorkingFolder() {
	async.RunE(func() error { return t.api.CleanWorkingFolder(t.repoID) }).
		Catch(func(err error) { t.ui.ShowErrorMessageBox("Failed to clean working folder:\n%s", err) })
}

func (t *repoVM) GetShownBranches(skipMaster bool) []api.Branch {
	branches, _ := t.api.GetBranches(
		api.GetBranchesReq{RepoID: t.repoID, IncludeOnlyShown: true, SkipMaster: skipMaster})
	return branches
}

func (t *repoVM) GetAllGitBranches() []api.Branch {
	return linq.Filter(t.GetAllBranches(), func(v api.Branch) bool { return v.IsGitBranch })
}

func (t *repoVM) GetAmbiguousBranches() []api.Branch {
	branches, _ := t.api.GetBranches(api.GetBranchesReq{RepoID: t.repoID, IncludeOnlyNotShown: false})
	return linq.Filter(branches, func(b api.Branch) bool { return b.AmbiguousTipId != "" })
}

func (t *repoVM) ShowBranch(name string, commitId string) {
	t.onRepoUpdatedFunc = func() { t.ScrollToBranch(name, commitId) }
	async.RunE(func() error { return t.api.ShowBranch(api.BranchName{RepoID: t.repoID, BranchName: name}) }).
		Catch(func(err error) { t.ui.ShowErrorMessageBox("Failed to show branch:\n%s\n%s", name, err) })
}

func (t *repoVM) ScrollToBranch(name string, commitId string) {
	t.ui.Post(func() {

		if commitId != "" {
			// Show specified commit at top
			_, i, ok := lo.FindIndexOf(t.repo.Commits, func(v api.Commit) bool { return v.ID == commitId })
			if !ok {
				return
			}

			t.repoViewer.ShowLineAtTop(i)
			return
		}

		// Show branch tip
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

func (t *repoVM) SetAsParentBranch(branchName, parentName string) {
	_ = t.api.SetAsParentBranch(api.SetParentReq{RepoID: t.repoID,
		BranchName: branchName, ParentName: parentName})
}

func (t *repoVM) UnsetAsParentBranch(name string) {
	_ = t.api.UnsetAsParentBranch(api.BranchName{RepoID: t.repoID, BranchName: name})
}

func (t *repoVM) HideBranch(name string) {
	_ = t.api.HideBranch(api.BranchName{RepoID: t.repoID, BranchName: name})
}

func (t *repoVM) SwitchToBranch(name string, displayName string) {
	async.RunE(func() error {
		return t.api.Checkout(t.repoID, name, displayName)
	}).
		Catch(func(err error) { t.ui.ShowErrorMessageBox("Failed to switch/checkout:\n%s\n%s", name, err) })
}

func (t *repoVM) PushBranch(name string) {
	p := t.ui.ShowProgress("")
	async.RunE(func() error { return t.api.PushBranch(t.repoID, name) }).
		Then(func(_ any) { p.Close() }).
		Catch(func(err error) {
			p.Close()
			t.ui.ShowErrorMessageBox("Failed to push:\n%s\n%s", name, err)
		})

}

func (t *repoVM) PushCurrentBranch() {
	current, ok := t.CurrentBranch()
	if !ok || !current.HasLocalOnly {
		return
	}
	t.startCommand(
		fmt.Sprintf("Pushing current branch:\n%s", current.Name),
		func() error { return t.api.PushBranch(t.repoID, current.Name) },
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
		func() error { return t.api.PullCurrentBranch(t.repoID) },
		func(err error) string { return fmt.Sprintf("Failed to pull/update:\n%s\n%s", current.Name, err) },
		nil)
}

func (t *repoVM) PullBranch(name string) {
	log.Infof("Pull branch %q", name)
	t.startCommand(
		fmt.Sprintf("Pull/Update branch:\n%s", name),
		func() error { return t.api.PullBranch(api.BranchName{RepoID: t.repoID, BranchName: name}) },
		func(err error) string { return fmt.Sprintf("Failed to pull/update:\n%s\n%s", name, err) },
		nil)
}

func (t *repoVM) MergeFromBranch(name string) {
	t.startCommand(
		fmt.Sprintf("Merging to Branch:\n%s", name),
		func() error { return t.api.MergeBranch(api.BranchName{RepoID: t.repoID, BranchName: name}) },
		func(err error) string { return fmt.Sprintf("Failed to merge:\n%s\n%s", name, err) },
		nil)
}

func (t *repoVM) MergeSquashFromBranch(name string) {
	t.startCommand(
		fmt.Sprintf("Merging to Branch:\n%s", name),
		func() error { return t.api.MergeSquashBranch(t.repoID, name) },
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
			progress.Close()
			if err != nil {
				msg := errorFunc(err)
				if msg != "" {
					t.ui.ShowErrorMessageBox(msg)
				}
			}
		})
	}()
}

func (t *repoVM) CreateBranch(name string) {
	t.startCommand(
		fmt.Sprintf("Creating Branch:\n%s", name),
		func() error {
			parent := t.repo.CurrentBranchName
			err := t.api.CreateBranch(api.BranchName{RepoID: t.repoID, BranchName: name, ParentName: parent})
			if err != nil {
				return err
			}

			err = t.api.PushBranch(t.repoID, name)
			if err != nil {
				return err
			}
			return err
		},
		func(err error) string { return fmt.Sprintf("Failed to create branch:\n%s\n%s", name, err) },
		func() { t.ShowBranch(name, "") })
}

func (t *repoVM) Clone(uri, path string) {
	progress := t.ui.ShowProgress(fmt.Sprintf("Cloning:\n%s\n%s", uri, path))
	t.api.CloneRepo(uri, path).
		Then(func(_ any) {
			progress.Close()
			log.Infof("Cloned %s into %s", uri, path)
			t.repoViewer.ShowRepo(path)
		}).
		Catch(func(err error) {
			progress.Close()
			t.ui.ShowErrorMessageBox("Failed to clone:\n%q into: \n%q\n%v", uri, path, err)
		})
}

func (t *repoVM) DeleteBranch(name string, isForced bool) {
	t.startCommand(
		fmt.Sprintf("Deleting Branch:\n%s", name),
		func() error {
			return t.api.DeleteBranch(t.repoID, name, isForced)
		},
		func(err error) string {
			if strings.Contains(err.Error(), "is not fully merged") {
				text := fmt.Sprintf("Branch %q is not fully merged.", name)
				text2 := "\n\nDo our want to force delete the branch?"
				msgBox := t.ui.MessageBox("Warning", cui.Yellow(text)+text2)
				msgBox.ShowCancel = true
				msgBox.OnOK = func() {
					t.DeleteBranch(name, true)
				}
				msgBox.Show()
				return ""
			} else {
				return fmt.Sprintf("Failed to delete:\n%s\n%s", name, err)
			}
		},
		nil)
}

func (t *repoVM) GetAmbiguousBranchBranchesMenuItems() []api.Branch {
	commit := t.repo.Commits[t.currentIndex]
	branch := t.repo.Branches[commit.BranchIndex]
	if !branch.IsAmbiguousBranch {
		return nil
	}

	branches, _ := t.api.GetAmbiguousBranchBranches(api.AmbiguousBranchBranchesReq{RepoID: t.repoID, CommitID: commit.ID})
	return branches
}
