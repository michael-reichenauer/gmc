package viewrepo

import (
	"context"
	"fmt"
	"github.com/michael-reichenauer/gmc/api"
	"github.com/michael-reichenauer/gmc/common/config"
	"github.com/michael-reichenauer/gmc/server/viewrepo/gitrepo"
	"github.com/michael-reichenauer/gmc/utils"
	"github.com/michael-reichenauer/gmc/utils/cui"
	"github.com/michael-reichenauer/gmc/utils/git"
	"github.com/michael-reichenauer/gmc/utils/log"
	"github.com/michael-reichenauer/gmc/utils/timer"
	"github.com/thoas/go-funk"
	"hash/fnv"
	"sort"
	"strings"
	"sync"
)

const (
	masterName       = "master"
	remoteMasterName = "origin/master"
)

type showRequest struct {
	branches   []string
	searchText string
}

type viewRepo struct {
	repoChanges chan api.RepoChange

	gitRepo       gitrepo.GitRepo
	configService *config.Service
	branchesGraph *branchesGraph

	showRequests       chan showRequest
	currentBranches    chan []string
	customBranchColors map[string]int
	ctx                context.Context
	cancel             context.CancelFunc
	viewRepo           *repo
	repoLock           sync.Mutex
}

func NewViewRepo(configService *config.Service, workingFolder string) api.Repo {
	ctx, cancel := context.WithCancel(context.Background())
	return &viewRepo{
		repoChanges:     make(chan api.RepoChange),
		showRequests:    make(chan showRequest),
		currentBranches: make(chan []string),
		branchesGraph:   newBranchesGraph(),
		gitRepo:         gitrepo.NewGitRepo(workingFolder),
		configService:   configService,
		ctx:             ctx,
		cancel:          cancel,
	}
}

func (t *viewRepo) storeViewRepo(viewRepo *repo) {
	t.repoLock.Lock()
	defer t.repoLock.Unlock()
	t.viewRepo = viewRepo
}

func (t *viewRepo) getViewRepo() *repo {
	t.repoLock.Lock()
	defer t.repoLock.Unlock()
	return t.viewRepo
}

func (t *viewRepo) StartMonitor() {
	go t.monitorViewModelRoutine(t.ctx)
}

func (t *viewRepo) Close() {
	t.cancel()
}

func (t *viewRepo) RepoChanges() chan api.RepoChange {
	return t.repoChanges
}

func (t *viewRepo) GetCurrentBranch() (api.Branch, bool) {
	viewRepo := t.getViewRepo()
	current, ok := viewRepo.gitRepo.CurrentBranch()
	if !ok {
		return api.Branch{}, false
	}

	for _, b := range viewRepo.Branches {
		if current.Name == b.name {
			return toBranch(b), true
		}
	}

	return toBranch(viewRepo.toBranch(current, 0)), true
}

func (t *viewRepo) TriggerRefreshModel() {
	log.Event("vms-refresh")
	t.gitRepo.TriggerManualRefresh()
}

func (t *viewRepo) GetCommitDiff(id string) (api.CommitDiff, error) {
	diff, err := t.gitRepo.GetCommitDiff(id)
	if err != nil {
		return api.CommitDiff{}, err
	}
	return api.CommitDiff{
		FileDiffs: toFileDiffs(diff.FileDiffs),
	}, nil
}

func (t *viewRepo) SwitchToBranch(name string, displayName string) error {
	if strings.HasPrefix(name, "origin/") {
		name = name[7:]
	}
	t.ShowBranch(name)
	// exist := false
	// for _, b := range repo.viewRepo.gitRepo.Branches {
	// 	if b.IsGitBranch && b.DisplayName == displayName {
	// 		exist = true
	// 		break
	// 	}
	// }
	// if !exist {
	// 	if err := t.gitRepo.CreateBranch(displayName); err != nil {
	// 		return err
	// 	}
	// }
	return t.gitRepo.SwitchToBranch(displayName)
}

func (t *viewRepo) Commit(Commit string) error {
	return t.gitRepo.Commit(Commit)
}

func (t *viewRepo) BranchColor(name string) cui.Color {
	if strings.HasPrefix(name, "multiple@") {
		return cui.CWhite
	}
	color, ok := t.customBranchColors[name]
	if ok {
		return cui.Color(color)
	}
	if name == "master" {
		return cui.CMagenta
	}
	if name == "develop" {
		return cui.CRedDk
	}

	h := fnv.New32a()
	h.Write([]byte(name))
	index := int(h.Sum32()) % len(branchColors)
	return branchColors[index]
}

func (t *viewRepo) GetCurrentNotShownBranch() (api.Branch, bool) {
	viewRepo := t.getViewRepo()
	current, ok := viewRepo.gitRepo.CurrentBranch()
	if !ok {
		return api.Branch{}, false
	}
	if containsBranch(viewRepo.Branches, current.Name) {
		return api.Branch{}, false
	}

	return toBranch(viewRepo.toBranch(current, 0)), true
}

func (t *viewRepo) GetAllBranches(skipShown bool) []api.Branch {
	viewRepo := t.getViewRepo()
	branches := t.getAllBranches(viewRepo, skipShown)
	sort.SliceStable(branches, func(i, j int) bool {
		return -1 == strings.Compare(branches[i].DisplayName, branches[j].DisplayName)
	})

	return branches
}

func (t *viewRepo) GetLatestBranches(skipShown bool) []api.Branch {
	viewRepo := t.getViewRepo()
	branches := t.getAllBranches(viewRepo, skipShown)
	sort.SliceStable(branches, func(i, j int) bool {
		return viewRepo.gitRepo.CommitByID(branches[i].TipID).AuthorTime.After(viewRepo.gitRepo.CommitByID(branches[j].TipID).AuthorTime)
	})

	return branches
}

func (t *viewRepo) GetCommitOpenInBranches(commitID string) []api.Branch {
	viewRepo := t.getViewRepo()
	c := viewRepo.gitRepo.CommitByID(commitID)
	var branches []*gitrepo.Branch

	if len(c.ParentIDs) > 1 {
		// commit has branch merged into this commit add it
		mergeParent := viewRepo.gitRepo.CommitByID(c.ParentIDs[1])
		branches = append(branches, mergeParent.Branch)
	}

	var bs []api.Branch
	for _, b := range branches {
		if nil != funk.Find(bs, func(bsb api.Branch) bool {
			return b.Name == bsb.Name || b.RemoteName == bsb.Name ||
				b.Name == bsb.RemoteName
		}) {
			// Skip duplicates
			continue
		}
		if nil != funk.Find(viewRepo.Branches, func(bsb *branch) bool {
			return b.Name == bsb.name
		}) {
			// Skip branches already shown
			continue
		}

		bs = append(bs, toBranch(viewRepo.toBranch(b, 0)))
	}

	return bs
}

func (t *viewRepo) GetCommitOpenOutBranches(commitID string) []api.Branch {
	viewRepo := t.getViewRepo()
	c := viewRepo.gitRepo.CommitByID(commitID)
	var branches []*gitrepo.Branch

	for _, ccId := range c.ChildIDs {
		// commit has children (i.e.), other branches have merged from this branch
		if ccId == git.UncommittedID {
			continue
		}
		cc := viewRepo.gitRepo.CommitByID(ccId)
		if cc.Branch.Name != c.Branch.Name {
			branches = append(branches, cc.Branch)
		}
	}

	for _, b := range viewRepo.gitRepo.Branches {
		if b.TipID == b.BottomID && b.BottomID == c.Id && b.ParentBranch.Name == c.Branch.Name {
			// empty branch with no own branch commit, (branch start)
			branches = append(branches, b)
		}
	}

	var bs []api.Branch
	for _, b := range branches {
		if nil != funk.Find(bs, func(bsb api.Branch) bool {
			return b.Name == bsb.Name || b.RemoteName == bsb.Name ||
				b.Name == bsb.RemoteName
		}) {
			// Skip duplicates
			continue
		}
		if nil != funk.Find(viewRepo.Branches, func(bsb *branch) bool {
			return b.Name == bsb.name
		}) {
			// Skip branches already shown
			continue
		}

		bs = append(bs, toBranch(viewRepo.toBranch(b, 0)))
	}

	return bs
}

//
func (t *viewRepo) GetShownBranches(skipMaster bool) []api.Branch {
	viewRepo := t.getViewRepo()
	var bs []api.Branch
	for _, b := range viewRepo.Branches {
		if skipMaster && (b.name == masterName || b.name == remoteMasterName) {
			// Do not support closing master branch
			continue
		}
		if b.isRemote && nil != funk.Find(viewRepo.Branches, func(bsb *branch) bool {
			return b.name == bsb.remoteName
		}) {
			// Skip remote if local exist
			continue
		}
		if nil != funk.Find(bs, func(bsb api.Branch) bool {
			return b.displayName == bsb.DisplayName
		}) {
			// Skip duplicates
			continue
		}

		bs = append(bs, toBranch(b))
	}
	return bs
}

func (t *viewRepo) ShowBranch(name string) {
	viewRepo := t.getViewRepo()
	branchNames := t.toBranchNames(viewRepo.Branches)

	branch, ok := viewRepo.gitRepo.BranchByName(name)
	if !ok {
		return
	}

	branchNames = t.addBranchWithAncestors(branchNames, branch)

	log.Event("vms-branches-open")
	t.showBranches(branchNames)
}

func (t *viewRepo) TriggerSearch(text string) {
	log.Eventf("vms-search", text)
	select {
	case t.showRequests <- showRequest{searchText: text}:
	default:
	}
}

//
func (t *viewRepo) HideBranch(name string) {
	viewRepo := t.getViewRepo()
	hideBranch, ok := funk.Find(viewRepo.Branches, func(b *branch) bool {
		return name == b.name
	}).(*branch)
	if !ok || hideBranch == nil {
		// No branch with that name
		return
	}
	if hideBranch.remoteName != "" {
		remoteBranch, ok := funk.Find(viewRepo.Branches, func(b *branch) bool {
			return hideBranch.remoteName == b.name
		}).(*branch)
		if ok && remoteBranch != nil {
			// The branch to hide has a remote branch, hiding that and the local branch
			// will be hidden as well
			hideBranch = remoteBranch
		}
	}

	var branchNames []string
	for _, b := range viewRepo.Branches {
		if b.name != hideBranch.name && !hideBranch.isAncestor(b) && b.remoteName != hideBranch.name {
			branchNames = append(branchNames, b.name)
		}
	}

	log.Event("vms-branches-close")
	t.showBranches(branchNames)
}

func (t *viewRepo) showBranches(branchIds []string) {
	log.Event("vms-load-repo")
	select {
	case t.showRequests <- showRequest{branches: branchIds}:
	default:
	}
}

func (t *viewRepo) monitorViewModelRoutine(ctx context.Context) {
	log.Infof("monitorViewModelRoutine start")
	defer log.Infof("View model monitor done")
	t.gitRepo.StartMonitor(ctx)

	var repo gitrepo.Repo
	var branchNames []string

	for {
		select {
		case change, ok := <-t.gitRepo.RepoChanges():
			if !ok {
				return
			}
			if change.IsStarting {
				log.Infof("Got repo start change ...")
				select {
				case t.repoChanges <- api.RepoChange{IsStarting: true}:
				case <-ctx.Done():
					return
				}
				break
			}

			log.Infof("Got repo change")

			if change.Error != nil {
				log.Event("vms-error-repo")
				select {
				case t.repoChanges <- api.RepoChange{Error: change.Error}:
				case <-ctx.Done():
					return
				}
				break
			}
			log.Event("vms-changed-repo")
			repo = change.Repo
			t.triggerFreshViewRepo(ctx, repo, branchNames)
		case request := <-t.showRequests:
			if request.searchText != "" {
				t.triggerSearchRepo(ctx, request.searchText, repo)
				break
			}
			log.Infof("Manual start change")
			if request.branches == nil {
				// A manual refresh, trigger new repo
				log.Infof("Manual refresh")
				t.gitRepo.TriggerManualRefresh()
				break
			}
			log.Infof("Refresh of %v", request.branches)
			branchNames = request.branches
			t.triggerFreshViewRepo(ctx, repo, branchNames)
		case names := <-t.currentBranches:
			branchNames = names
		case <-ctx.Done():
			return
		}
	}
}

func (t *viewRepo) triggerSearchRepo(ctx context.Context, searchText string, repo gitrepo.Repo) {
	log.Infof("triggerSearchRepo")
	go func() {
		vRepo := t.getSearchModel(repo, searchText)
		repoChange := api.RepoChange{SearchText: searchText, ViewRepo: toViewRepo(vRepo)}
		select {
		case t.repoChanges <- repoChange:
		case <-ctx.Done():
		}
	}()
}

func (t *viewRepo) triggerFreshViewRepo(ctx context.Context, repo gitrepo.Repo, branchNames []string) {
	log.Infof("triggerFreshViewRepo")
	go func() {
		vRepo := t.getViewModel(repo, branchNames)
		t.storeViewRepo(vRepo)
		vr := toViewRepo(vRepo)
		repoChange := api.RepoChange{ViewRepo: vr}
		select {
		case t.repoChanges <- repoChange:
			t.storeShownBranchesInConfig(branchNames)
		case <-ctx.Done():
		}
		branchNames := t.getBranchNames(vRepo)
		select {
		case t.currentBranches <- branchNames:
		case <-ctx.Done():
		}
	}()
}

func (t *viewRepo) storeShownBranchesInConfig(branchNames []string) {
	t.configService.SetRepo(t.gitRepo.RepoPath(), func(r *config.Repo) {
		r.ShownBranches = branchNames
	})
}

func (t *viewRepo) getSearchModel(grepo gitrepo.Repo, searchText string) *repo {
	log.Infof("getSearchModel")
	ti := timer.Start()
	repo := newRepo()
	repo.gitRepo = grepo
	repo.WorkingFolder = grepo.RepoPath
	currentBranch, ok := grepo.CurrentBranch()
	repo.UncommittedChanges = grepo.Status.AllChanges()
	repo.Conflicts = grepo.Status.Conflicted

	if ok {
		repo.CurrentBranchName = currentBranch.Name
	}

	for _, b := range grepo.Branches {
		repo.addBranch(b)
	}

	for _, c := range grepo.SearchCommits(searchText) {
		repo.addSearchCommit(c)
	}
	t.addTags(repo, grepo.Tags)
	log.Infof("done, %s", ti)
	return repo
}

func (t *viewRepo) getViewModel(grepo gitrepo.Repo, branchNames []string) *repo {
	log.Infof("getViewModel")
	ti := timer.Start()
	repo := newRepo()
	repo.gitRepo = grepo
	repo.WorkingFolder = grepo.RepoPath
	repo.UncommittedChanges = grepo.Status.AllChanges()
	repo.Conflicts = grepo.Status.Conflicted
	repo.MergeMessage = grepo.Status.MergeMessage

	branches := t.getGitRepoBranches(branchNames, grepo)
	for _, b := range branches {
		repo.addBranch(b)
	}

	currentBranch, ok := grepo.CurrentBranch()
	if ok {
		repo.CurrentBranchName = currentBranch.Name
	}

	if !grepo.Status.OK() {
		repo.addVirtualStatusCommit(grepo)
	}
	for _, c := range grepo.Commits {
		repo.addGitCommit(c)
	}

	t.addTags(repo, grepo.Tags)

	t.adjustCurrentBranchIfStatus(repo)
	t.setBranchParentChildRelations(repo)
	t.setParentChildRelations(repo)
	t.setAheadBehind(repo)
	t.setBranchColors(repo)

	// Draw branch lines
	t.branchesGraph.drawBranchLines(repo)

	// Draw branch connector lines
	t.branchesGraph.drawConnectorLines(repo)
	log.Infof("getViewModel done, %s", ti)
	return repo
}

func (t *viewRepo) adjustCurrentBranchIfStatus(repo *repo) {
	if len(repo.Commits) < 2 || repo.Commits[0].ID != git.UncommittedID || repo.CurrentCommit == nil {
		return
	}

	// Link status commit with current commit and adjust branch
	statusCommit := repo.Commits[0]
	current := repo.CurrentCommit
	statusCommit.Parent = current
	current.ChildIDs = append([]string{statusCommit.ID}, current.ChildIDs...)
	statusCommit.Branch.tip = statusCommit
	statusCommit.Branch.tipId = statusCommit.ID

	if statusCommit.Branch.name != current.Branch.name {
		// Status commit is first local commit, lets adjust branch
		statusCommit.Branch.bottom = statusCommit
		statusCommit.Branch.bottomId = statusCommit.ID
	}
}

func (t *viewRepo) getAllBranches(viewRepo *repo, skipShown bool) []api.Branch {
	var branches []api.Branch
	for _, b := range viewRepo.gitRepo.Branches {
		if containsDisplayNameBranch(branches, b.DisplayName) {
			continue
		}
		if skipShown && containsBranch(viewRepo.Branches, b.Name) {
			continue
		}
		branches = append(branches, toBranch(viewRepo.toBranch(b, 0)))
	}

	return branches
}

func containsDisplayNameBranch(branches []api.Branch, displayName string) bool {
	for _, b := range branches {
		if displayName == b.DisplayName {
			return true
		}
	}
	return false
}

//
func containsBranch(branches []*branch, name string) bool {
	for _, b := range branches {
		if name == b.name {
			return true
		}
	}
	return false
}
func (t *viewRepo) setBranchParentChildRelations(repo *repo) {
	for _, b := range repo.Branches {
		b.tip = repo.commitById[b.tipId]
		b.bottom = repo.commitById[b.bottomId]
		if b.parentBranchName != "" {
			b.parentBranch = repo.BranchByName(b.parentBranchName)
		}
	}
}

func (t *viewRepo) setParentChildRelations(repo *repo) {
	for _, c := range repo.Commits {
		if len(c.ParentIDs) > 0 {
			// Commit has a parent
			c.Parent = repo.commitById[c.ParentIDs[0]]
			if len(c.ParentIDs) > 1 {
				// Merge parent can be nil if not the merge parent branch is included in the repo view model as well
				c.MergeParent = repo.commitById[c.ParentIDs[1]]
				if c.MergeParent == nil {
					// commit has a merge parent, that is not visible, mark the commit as expandable
					c.More.Set(api.MoreMergeIn)
				}
			}
		}

		for _, childID := range c.ChildIDs {
			if _, ok := repo.commitById[childID]; !ok {
				// commit has a child, that is not visible, mark the commit as expandable
				c.More.Set(api.MoreBranchOut)
			}
		}

		if c.More == api.MoreNone {
			for _, branchName := range c.BranchTips {
				if !repo.containsBranchName(branchName) {
					// Some not shown branch tip att this commit
					c.More.Set(api.MoreBranchOut)
					break
				}
			}
		}
	}
}

func (t *viewRepo) toBranchNames(branches []*branch) []string {
	var ids []string
	for _, b := range branches {
		ids = append(ids, b.name)
	}
	return ids
}

func (t *viewRepo) getBranchNames(repo *repo) []string {
	var names []string
	for _, b := range repo.Branches {
		names = append(names, b.name)
	}
	return names
}
func (t *viewRepo) setAheadBehind(repo *repo) {
	for _, b := range repo.Branches {
		if b.isRemote && b.localName != "" {
			t.setIsRemoteOnlyCommits(repo, b)
		} else if !b.isRemote && b.remoteName != "" {
			t.setLocalOnlyCommits(repo, b)
		}
	}
}

func (t *viewRepo) setIsRemoteOnlyCommits(repo *repo, b *branch) {
	localBranch := repo.tryGetBranchByName(b.localName)
	localTip := t.tryGetRealLocalTip(localBranch)
	localBase := localBranch.bottom.Parent

	count := 0
	for c := b.tip; c != nil && c.Branch == b && count < 50; c = c.Parent {
		count++
		if localTip != nil && localTip == c {
			// Local branch tip on same commit as remote branch tip (i.e. synced)
			break
		}
		if localBase != nil && localBase == c {
			// Local branch base on same commit as remote branch tip (i.e. synced from this point)
			break
		}
		if c.MergeParent != nil && c.MergeParent.Branch.name == b.localName {
			break
		}
		c.IsRemoteOnly = true
		b.HasRemoteOnly = true
	}
	if b.HasRemoteOnly {
		bl := repo.BranchByName(b.localName)
		if bl != nil {
			bl.HasRemoteOnly = true
		}
	}
}

func (t *viewRepo) setLocalOnlyCommits(repo *repo, b *branch) {
	count := 0
	for c := b.tip; c != nil && c.Branch == b && count < 50; c = c.Parent {
		if c.ID == git.UncommittedID {
			continue
		}
		count++
		c.IsLocalOnly = true
		b.HasLocalOnly = true
	}
	if b.HasLocalOnly {
		br := repo.BranchByName(b.remoteName)
		if br != nil {
			br.HasLocalOnly = true
		}
	}
}

func (t *viewRepo) tryGetRealLocalTip(localBranch *branch) *commit {
	var localTip *commit
	if localBranch != nil {
		localTip = localBranch.tip
		if localTip.ID == git.UncommittedID {
			localTip = localTip.Parent
		}
	}
	return localTip
}

func (t *viewRepo) isSynced(b *branch) bool {
	return b.isRemote && b.localName == "" || !b.isRemote && b.remoteName == ""
}

func (t *viewRepo) PushBranch(name string) error {
	return t.gitRepo.PushBranch(name)
}

func (t *viewRepo) PullBranch() error {
	return t.gitRepo.PullBranch()
}

func (t *viewRepo) MergeBranch(name string) error {
	return t.gitRepo.MergeBranch(name)
}

func (t *viewRepo) CreateBranch(name string) error {
	return t.gitRepo.CreateBranch(name)
}

func (t *viewRepo) DeleteBranch(name string) error {
	viewRepo := t.getViewRepo()
	branch, ok := viewRepo.gitRepo.BranchByName(name)
	if !ok {
		return fmt.Errorf("unknown git branch %q", name)
	}
	if !branch.IsGitBranch {
		return fmt.Errorf("not a git branch %q", name)
	}
	if utils.StringsContains(gitrepo.DefaultBranchPriority, name) {
		return fmt.Errorf("branch is protected %q", name)
	}

	var localBranch *gitrepo.Branch
	var remoteBranch *gitrepo.Branch

	if branch.IsRemote {
		// Remote branch, check if there is a corresponding local branch
		remoteBranch = branch
		if branch.LocalName != "" {
			if b, ok := viewRepo.gitRepo.BranchByName(branch.LocalName); ok {
				localBranch = b
			}
		}
	}

	if !branch.IsRemote {
		// Local branch, check if there is a corresponding remote branch
		localBranch = branch
		if branch.RemoteName != "" {
			if b, ok := viewRepo.gitRepo.BranchByName(branch.RemoteName); ok {
				remoteBranch = b
			}
		}
	}

	if remoteBranch != nil {
		// Deleting remote branch
		err := t.gitRepo.DeleteRemoteBranch(remoteBranch.Name)
		if err != nil {
			return err
		}
	}

	if localBranch != nil {
		// Deleting local branch
		err := t.gitRepo.DeleteLocalBranch(localBranch.Name)
		if err != nil {
			return err
		}
	}
	return nil
}

func (t *viewRepo) addTags(repo *repo, tags []gitrepo.Tag) {
	for _, tag := range tags {
		c, ok := repo.commitById[tag.CommitID]
		if !ok {
			continue
		}
		c.Tags = append(c.Tags, tag.TagName)
	}
}
