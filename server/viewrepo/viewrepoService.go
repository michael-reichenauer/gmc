package viewrepo

import (
	"context"
	"fmt"
	"hash/fnv"
	"sort"
	"strings"
	"sync"

	"github.com/imkira/go-observer"
	"github.com/michael-reichenauer/gmc/api"
	"github.com/michael-reichenauer/gmc/common/config"
	"github.com/michael-reichenauer/gmc/server/viewrepo/augmented"
	"github.com/michael-reichenauer/gmc/utils"
	"github.com/michael-reichenauer/gmc/utils/cui"
	"github.com/michael-reichenauer/gmc/utils/git"
	"github.com/michael-reichenauer/gmc/utils/log"
	"github.com/michael-reichenauer/gmc/utils/timer"
	"github.com/samber/lo"
	"github.com/thoas/go-funk"
)

const (
	masterName       = "master"
	remoteMasterName = "origin/master"
	mainName         = "main"
	remoteMainName   = "origin/main"
)

type showRequest struct {
	branches   []string
	searchText string
}

type ViewRepoService struct {
	changes observer.Property

	augmentedRepo augmented.RepoService
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

func NewViewRepoService(configService *config.Service, rootPath string) *ViewRepoService {
	ctx, cancel := context.WithCancel(context.Background())

	return &ViewRepoService{
		changes:         observer.NewProperty(nil),
		showRequests:    make(chan showRequest),
		currentBranches: make(chan []string),
		branchesGraph:   newBranchesGraph(),
		augmentedRepo:   augmented.NewRepoService(configService, rootPath),
		configService:   configService,
		ctx:             ctx,
		cancel:          cancel,
	}
}
func (t *ViewRepoService) ObserveChanges() observer.Stream {
	return t.changes.Observe()
}

func (t *ViewRepoService) storeViewRepo(viewRepo *repo) {
	t.repoLock.Lock()
	defer t.repoLock.Unlock()
	t.viewRepo = viewRepo
}

func (t *ViewRepoService) getViewRepo() *repo {
	t.repoLock.Lock()
	defer t.repoLock.Unlock()
	return t.viewRepo
}

func (t *ViewRepoService) StartMonitor() {
	go t.monitorViewModelRoutine(t.ctx)
}

func (t *ViewRepoService) CloseRepo() {
	t.cancel()
	//close(t.repoChangesIn)
}

func (t *ViewRepoService) TriggerRefreshModel() {
	log.Event("vms-refresh")
	t.augmentedRepo.TriggerManualRefresh()
}

func (t *ViewRepoService) GetCommitDiff(id string) (api.CommitDiff, error) {
	diff, err := t.augmentedRepo.GetCommitDiff(id)
	if err != nil {
		return api.CommitDiff{}, err
	}
	return ToApiCommitDiff(diff), nil
}

func (t *ViewRepoService) GetCommitDetails(id string) (api.CommitDetailsRsp, error) {
	c, ok := t.viewRepo.CommitById(id)
	if !ok {
		return api.CommitDetailsRsp{}, fmt.Errorf("unknown commit %q", id)
	}
	diff, err := t.augmentedRepo.GetCommitDiff(id)
	if err != nil {
		return api.CommitDetailsRsp{}, err
	}

	files := lo.Map(diff.FileDiffs, func(v git.FileDiff, _ int) string {
		return v.PathBefore
	})

	return api.CommitDetailsRsp{Id: c.ID, BranchName: c.Branch.displayName, Message: c.Message, Files: files}, nil
}

func (t *ViewRepoService) SwitchToBranch(name string, displayName string) error {
	name = strings.TrimPrefix(name, "origin/")
	t.ShowBranch(name)
	// exist := false
	// for _, b := range repo.viewRepo.augmentedRepo.Branches {
	// 	if b.IsGitBranch && b.DisplayName == displayName {
	// 		exist = true
	// 		break
	// 	}
	// }
	// if !exist {
	// 	if err := t.augmentedRepo.CreateBranch(displayName); err != nil {
	// 		return err
	// 	}
	// }
	return t.augmentedRepo.SwitchToBranch(displayName)
}

func (t *ViewRepoService) Commit(Commit string) error {
	return t.augmentedRepo.Commit(Commit)
}

func (t *ViewRepoService) BranchColor(branch *branch) cui.Color {
	log.Infof("Branch %q", branch.name)
	if branch.parentBranch == nil {
		// branch has no parent or parent is remote of this branch, lets use it
		log.Infof("No parent branch %q", branch.name)
		return t.branchNameColor(branch.displayName, 0)
	}

	if branch.remoteName == branch.parentBranch.name {
		// Parent is remote of this branch, lets use parent color
		log.Infof("Parent is remote branch %q", branch.name)
		return t.BranchColor(branch.parentBranch)
	}

	color := t.branchNameColor(branch.displayName, 0)
	parentColor := t.branchNameColor(branch.parentBranch.displayName, 0)

	if color == parentColor {
		// branch got same color as parent, lets change branch color
		color = t.branchNameColor(branch.displayName, 1)
	}

	log.Infof("Branch %q %q %d %d", branch.name, branch.parentBranch.displayName, color, parentColor)

	return color
}

func (t *ViewRepoService) branchNameColor(name string, addIndex int) cui.Color {
	if strings.HasPrefix(name, "ambiguous@") {
		return cui.CWhite
	}
	color, ok := t.customBranchColors[name]
	if ok {
		return cui.Color(color)
	}
	if name == masterName || name == mainName {
		return cui.CMagenta
	}

	h := fnv.New32a()
	h.Write([]byte(name))
	index := int(h.Sum32()+uint32(addIndex)) % len(branchColors)
	return branchColors[index]
}

func (t *ViewRepoService) GetBranches(args api.GetBranchesReq) []api.Branch {
	branches := []api.Branch{}
	if args.IncludeOnlyCommitBranches != "" {
		return append(branches, t.getCommitBranches(args.IncludeOnlyCommitBranches)...)
	} else if args.IncludeOnlyCurrent {
		if b, ok := t.getCurrentBranch(args.IncludeOnlyNotShown); ok {
			branches = append(branches, b)
		}
		return branches
	} else if args.IncludeOnlyShown {
		return t.getShownBranches(args.SkipMaster)
	}

	return t.getAllBranchesC(args.IncludeOnlyNotShown, args.SortOnLatest)
}

func (t *ViewRepoService) GetAmbiguousBranchBranches(args api.AmbiguousBranchBranchesReq) []api.Branch {
	branches := []api.Branch{}

	viewRepo := t.getViewRepo()

	commit, ok := viewRepo.CommitById(args.CommitID)
	if !ok {
		return nil
	}

	for _, name := range commit.Branch.ambiguousBranchNames {
		branch, ok := viewRepo.augmentedRepo.BranchByName((name))
		if !ok {
			continue
		}

		// found := false
		// for _, bbb := range viewRepo.Branches {
		// 	if name == bbb.name {
		// 		found = true
		// 		break
		// 	}
		// }
		// if found {
		// 	continue
		// }

		branches = append(branches, toApiBranch(viewRepo.toBranch(branch, 0)))
	}

	return branches
}

func (t *ViewRepoService) getAllBranchesC(skipShown bool, sortOnLatest bool) []api.Branch {
	viewRepo := t.getViewRepo()
	branches := t.getAllBranches(viewRepo, skipShown)
	if sortOnLatest {
		sort.SliceStable(branches, func(i, j int) bool {
			return viewRepo.augmentedRepo.CommitByID(branches[i].TipID).AuthorTime.After(viewRepo.augmentedRepo.CommitByID(branches[j].TipID).AuthorTime)
		})
	} else {
		sort.SliceStable(branches, func(i, j int) bool {
			return strings.Compare(branches[i].DisplayName, branches[j].DisplayName) == -1
		})
	}

	return branches
}

func (t *ViewRepoService) getCurrentBranch(includeOnlyNotShown bool) (api.Branch, bool) {
	viewRepo := t.getViewRepo()
	current, ok := viewRepo.augmentedRepo.CurrentBranch()
	if !ok {
		return api.Branch{}, false
	}
	if includeOnlyNotShown && containsBranch(viewRepo.Branches, current.Name) {
		return api.Branch{}, false
	}

	for _, b := range viewRepo.Branches {
		if current.Name == b.name {
			return toApiBranch(b), true
		}
	}

	return toApiBranch(viewRepo.toBranch(current, 0)), true
}

func (t *ViewRepoService) getCommitBranches(commitID string) []api.Branch {
	branches := t.getCommitOpenInBranches(commitID)
	branches = append(branches, t.getCommitOpenOutBranches(commitID)...)
	return branches
}

func (t *ViewRepoService) getCommitOpenInBranches(commitID string) []api.Branch {
	viewRepo := t.getViewRepo()
	c := viewRepo.augmentedRepo.CommitByID(commitID)
	var branches []*augmented.Branch

	if len(c.ParentIDs) > 1 {
		// commit has branch merged into this commit add it
		mergeParent := viewRepo.augmentedRepo.CommitByID(c.ParentIDs[1])
		branches = append(branches, mergeParent.Branch)
	}

	bs := []api.Branch{}
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
		branch := toApiBranch(viewRepo.toBranch(b, 0))
		branch.IsIn = true
		bs = append(bs, branch)
	}

	return bs
}

func (t *ViewRepoService) getCommitOpenOutBranches(commitID string) []api.Branch {
	viewRepo := t.getViewRepo()
	c := viewRepo.augmentedRepo.CommitByID(commitID)
	var branches []*augmented.Branch

	for _, ccId := range c.ChildIDs {
		// commit has children (i.e.), other branches have merged from this branch
		if ccId == git.UncommittedID {
			continue
		}
		cc := viewRepo.augmentedRepo.CommitByID(ccId)
		if cc.Branch.Name != c.Branch.Name {
			branches = append(branches, cc.Branch)
		}
	}

	for _, b := range viewRepo.augmentedRepo.Branches {
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
		branch := toApiBranch(viewRepo.toBranch(b, 0))
		branch.IsOut = true
		bs = append(bs, branch)
	}

	return bs
}

func (t *ViewRepoService) getShownBranches(skipMaster bool) []api.Branch {
	viewRepo := t.getViewRepo()
	bs := []api.Branch{}
	for _, b := range viewRepo.Branches {
		if skipMaster &&
			(b.name == masterName || b.name == remoteMasterName ||
				b.name == mainName || b.name == remoteMainName) {
			// Do not support closing main branch
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

		bs = append(bs, toApiBranch(b))
	}
	return bs
}

func (t *ViewRepoService) ShowBranch(name string) {
	viewRepo := t.getViewRepo()
	branchNames := t.toBranchNames(viewRepo.Branches)

	branch, ok := viewRepo.augmentedRepo.BranchByName(name)
	if !ok {
		return
	}

	branchNames = t.addBranchWithAncestors(branchNames, branch)
	if branch.IsAmbiguousBranch {
		// For ambiguous branches, lets show its children branches as well to determine which
		// child should be the parent
		for _, b := range branch.AmbiguousBranches {
			if !lo.Contains(branchNames, b.Name) {
				branchNames = append(branchNames, b.Name)
			}
		}
	}

	log.Event("vms-branches-open")
	t.showBranches(branchNames)
}

func (t *ViewRepoService) TriggerSearch(text string) {
	log.Eventf("vms-search", text)
	select {
	case t.showRequests <- showRequest{searchText: text}:
	default:
	}
}

func (t *ViewRepoService) HideBranch(name string) {
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

func (t *ViewRepoService) showBranches(branchIds []string) {
	log.Event("vms-load-repo")
	select {
	case t.showRequests <- showRequest{branches: branchIds}:
	default:
	}
}

func (t *ViewRepoService) monitorViewModelRoutine(ctx context.Context) {
	log.Infof("monitorViewModelRoutine start")
	defer log.Infof("View model monitor done")
	t.augmentedRepo.StartMonitor(ctx)

	var repo augmented.Repo
	var branchNames []string

	for {
		select {
		case change, ok := <-t.augmentedRepo.RepoChanges():
			if !ok {
				return
			}
			if change.IsStarting {
				log.Infof("Send repo start change ...")
				select {
				case <-ctx.Done():
					return
				default:
					t.changes.Update(api.RepoChange{IsStarting: true})
				}
				break
			}

			log.Infof("Got repo change")

			if change.Error != nil {
				log.Event("vms-error-repo")
				log.Infof("Send repo error event ...")
				select {
				case <-ctx.Done():
					return
				default:
					t.changes.Update(api.RepoChange{Error: change.Error})
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
				t.augmentedRepo.TriggerManualRefresh()
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

func (t *ViewRepoService) triggerSearchRepo(ctx context.Context, searchText string, repo augmented.Repo) {
	log.Infof("triggerSearchRepo")
	go func() {
		vRepo := t.getSearchModel(repo, searchText)
		repoChange := api.RepoChange{SearchText: searchText, ViewRepo: toApiRepo(vRepo)}
		log.Infof("Send new search repo event ...")
		select {
		case <-ctx.Done():
		default:
			t.changes.Update(repoChange)
		}
	}()
}

func (t *ViewRepoService) triggerFreshViewRepo(ctx context.Context, repo augmented.Repo, branchNames []string) {
	log.Infof("triggerFreshViewRepo")
	go func() {
		vRepo := t.GetViewModel(repo, branchNames)
		t.storeViewRepo(vRepo)
		vr := toApiRepo(vRepo)
		repoChange := api.RepoChange{ViewRepo: vr}
		log.Infof("Send new view repo event ...")
		select {
		case <-ctx.Done():
		default:
			t.changes.Update(repoChange)
			t.storeShownBranchesInConfig(branchNames)
		}
		branchNames := t.getBranchNames(vRepo)
		select {
		case t.currentBranches <- branchNames:
		case <-ctx.Done():
		}
	}()
}

func (t *ViewRepoService) storeShownBranchesInConfig(branchNames []string) {
	t.configService.SetRepo(t.augmentedRepo.RepoPath(), func(r *config.Repo) {
		r.ShownBranches = branchNames
	})
}

func (t *ViewRepoService) getSearchModel(gRepo augmented.Repo, searchText string) *repo {
	log.Infof("getSearchModel")
	ti := timer.Start()
	repo := newRepo()
	repo.augmentedRepo = gRepo
	repo.WorkingFolder = gRepo.RepoPath
	currentBranch, ok := gRepo.CurrentBranch()
	repo.UncommittedChanges = gRepo.Status.AllChanges()
	repo.Conflicts = gRepo.Status.Conflicted

	if ok {
		repo.CurrentBranchName = currentBranch.Name
	}

	for _, b := range gRepo.Branches {
		repo.addBranch(b)
	}

	for _, c := range gRepo.SearchCommits(searchText) {
		repo.addSearchCommit(c)
	}
	t.addTags(repo, gRepo.Tags)
	log.Infof("done, %s", ti)
	return repo
}

func (t *ViewRepoService) GetViewModel(augRepo augmented.Repo, branchNames []string) *repo {
	log.Infof("getViewModel")
	ti := timer.Start()
	repo := newRepo()
	repo.augmentedRepo = augRepo
	repo.WorkingFolder = augRepo.RepoPath
	repo.UncommittedChanges = augRepo.Status.AllChanges()
	repo.Conflicts = augRepo.Status.Conflicted
	repo.MergeMessage = augRepo.Status.MergeMessage

	branches := t.getAugmentedBranches(branchNames, augRepo)
	for _, b := range branches {
		repo.addBranch(b)
	}

	currentBranch, ok := augRepo.CurrentBranch()
	if ok {
		repo.CurrentBranchName = currentBranch.Name
	}

	if !augRepo.Status.OK() {
		repo.addVirtualStatusCommit(augRepo)
	}
	for _, c := range augRepo.Commits {
		repo.addGitCommit(c)
	}

	t.addTags(repo, augRepo.Tags)

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

func (t *ViewRepoService) adjustCurrentBranchIfStatus(repo *repo) {
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

func (t *ViewRepoService) getAllBranches(viewRepo *repo, skipShown bool) []api.Branch {
	branches := []api.Branch{}
	for _, b := range viewRepo.augmentedRepo.Branches {
		if containsDisplayNameBranch(branches, b.DisplayName) {
			continue
		}
		if skipShown && containsBranch(viewRepo.Branches, b.Name) {
			continue
		}
		branches = append(branches, toApiBranch(viewRepo.toBranch(b, 0)))
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

func containsBranch(branches []*branch, name string) bool {
	for _, b := range branches {
		if name == b.name {
			return true
		}
	}
	return false
}
func (t *ViewRepoService) setBranchParentChildRelations(repo *repo) {
	for _, b := range repo.Branches {
		b.tip = repo.commitById[b.tipId]
		b.bottom = repo.commitById[b.bottomId]
		if b.parentBranchName != "" {
			b.parentBranch = repo.BranchByName(b.parentBranchName)
		}
	}
}

func (t *ViewRepoService) setParentChildRelations(repo *repo) {
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

func (t *ViewRepoService) toBranchNames(branches []*branch) []string {
	var ids []string
	for _, b := range branches {
		ids = append(ids, b.name)
	}
	return ids
}

func (t *ViewRepoService) getBranchNames(repo *repo) []string {
	var names []string
	for _, b := range repo.Branches {
		names = append(names, b.name)
	}
	return names
}
func (t *ViewRepoService) setAheadBehind(repo *repo) {
	for _, b := range repo.Branches {
		if b.isRemote && b.localName != "" {
			t.setIsRemoteOnlyCommits(repo, b)
		} else if !b.isRemote && b.remoteName != "" {
			t.setLocalOnlyCommits(repo, b)
		}
	}
}

func (t *ViewRepoService) setIsRemoteOnlyCommits(repo *repo, b *branch) {
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

func (t *ViewRepoService) setLocalOnlyCommits(repo *repo, b *branch) {
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

func (t *ViewRepoService) tryGetRealLocalTip(localBranch *branch) *commit {
	var localTip *commit
	if localBranch != nil {
		localTip = localBranch.tip
		if localTip.ID == git.UncommittedID {
			localTip = localTip.Parent
		}
	}
	return localTip
}

func (t *ViewRepoService) isSynced(b *branch) bool {
	return b.isRemote && b.localName == "" || !b.isRemote && b.remoteName == ""
}

func (t *ViewRepoService) PushBranch(name string) error {
	return t.augmentedRepo.PushBranch(name)
}

func (t *ViewRepoService) PullCurrentBranch() error {
	return t.augmentedRepo.PullCurrentBranch()
}

func (t *ViewRepoService) PullBranch(name string) error {
	return t.augmentedRepo.PullBranch(name)
}

func (t *ViewRepoService) MergeBranch(name string) error {
	return t.augmentedRepo.MergeBranch(name)
}

func (t *ViewRepoService) CreateBranch(name string) error {
	return t.augmentedRepo.CreateBranch(name)
}

func (t *ViewRepoService) DeleteBranch(name string) error {
	viewRepo := t.getViewRepo()
	branch, ok := viewRepo.augmentedRepo.BranchByName(name)
	if !ok {
		return fmt.Errorf("unknown git branch %q", name)
	}
	if !branch.IsGitBranch {
		return fmt.Errorf("not a git branch %q", name)
	}
	if utils.StringsContains(augmented.DefaultBranchPriority, name) {
		return fmt.Errorf("branch is protected %q", name)
	}

	var localBranch *augmented.Branch
	var remoteBranch *augmented.Branch

	if branch.IsRemote {
		// Remote branch, check if there is a corresponding local branch
		remoteBranch = branch
		if branch.LocalName != "" {
			if b, ok := viewRepo.augmentedRepo.BranchByName(branch.LocalName); ok {
				localBranch = b
			}
		}
	}

	if !branch.IsRemote {
		// Local branch, check if there is a corresponding remote branch
		localBranch = branch
		if branch.RemoteName != "" {
			if b, ok := viewRepo.augmentedRepo.BranchByName(branch.RemoteName); ok {
				remoteBranch = b
			}
		}
	}

	if remoteBranch != nil {
		// Deleting remote branch
		err := t.augmentedRepo.DeleteRemoteBranch(remoteBranch.Name)
		if err != nil {
			return err
		}
	}

	if localBranch != nil {
		// Deleting local branch
		err := t.augmentedRepo.DeleteLocalBranch(localBranch.Name)
		if err != nil {
			return err
		}
	}
	return nil
}

func (t *ViewRepoService) SetAsParentBranch(name string) error {
	viewRepo := t.getViewRepo()
	b, ok := viewRepo.augmentedRepo.BranchByName(name)
	if !ok {
		return fmt.Errorf("unknown branch %q", name)
	}

	if b.ParentBranch == nil {
		return fmt.Errorf("branch has no parent branch %q", name)
	}

	if !b.ParentBranch.IsAmbiguousBranch {
		return fmt.Errorf("parent branch is not a ambiguous branch %q", b.ParentBranch.Name)
	}

	parentName := name
	otherChildren := lo.Filter(b.ParentBranch.AmbiguousBranches, func(v *augmented.Branch, _ int) bool {
		return v.Name != name
	})
	childNames := lo.Map(otherChildren, func(v *augmented.Branch, _ int) string {
		return v.Name
	})

	t.configService.SetRepo(viewRepo.WorkingFolder, func(r *config.Repo) {
		parentChildrenNames, ok := r.BranchesChildren[parentName]
		if !ok {
			parentChildrenNames = []string{}
			r.BranchesChildren[parentName] = parentChildrenNames
		}

		for _, childName := range childNames {
			// Ensure parent branch is not a child of any of the children
			childChildrenNames, ok := r.BranchesChildren[childName]
			if ok {
				childChildrenNames = lo.Filter(childChildrenNames, func(v string, _ int) bool {
					return v != parentName
				})
				r.BranchesChildren[childName] = childChildrenNames
			}

			// Add child name as child to parent
			if !lo.Contains(parentChildrenNames, childName) {
				parentChildrenNames = append(parentChildrenNames, childName)
				r.BranchesChildren[parentName] = parentChildrenNames
			}
		}
	})

	t.TriggerRefreshModel()
	return nil
}

func (t *ViewRepoService) UnsetAsParentBranch(name string) error {
	viewRepo := t.getViewRepo()

	t.configService.SetRepo(viewRepo.WorkingFolder, func(r *config.Repo) {
		_, ok := r.BranchesChildren[name]
		if !ok {
			return
		}
		delete(r.BranchesChildren, name)
	})

	t.TriggerRefreshModel()
	return nil
}

func (t *ViewRepoService) addTags(repo *repo, tags []augmented.Tag) {
	for _, tag := range tags {
		c, ok := repo.commitById[tag.CommitID]
		if !ok {
			continue
		}
		c.Tags = append(c.Tags, tag.TagName)
	}
}

func isMainBranch(name string) bool {
	return name == masterName || name == remoteMasterName || name == mainName || name == remoteMainName
}
