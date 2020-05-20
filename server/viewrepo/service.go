package viewrepo

import (
	"context"
	"fmt"
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
)

const (
	masterName       = "master"
	remoteMasterName = "origin/master"
)

var branchColors = []cui.Color{
	cui.CRed,
	cui.CBlue,
	cui.CYellow,
	cui.CGreen,
	cui.CCyan,
	cui.CRedDk,
	cui.CGreenDk,
	cui.CYellowDk,
	//ui.CBlueDk,
	cui.CMagenta,
	cui.CMagentaDk,
	cui.CCyanDk,
}

type Status struct {
	AllChanges int
	GraphWidth int
}

type RepoChange struct {
	IsStarting bool
	ViewRepo   Repo
	SearchText string
	Error      error
}

type ShowRequest struct {
	branches   []string
	searchText string
}

type ViewRepo struct {
	RepoChanges chan RepoChange

	gitRepo       gitrepo.GitRepo
	configService *config.Service
	branchesGraph *branchesGraph

	showRequests       chan ShowRequest
	currentBranches    chan []string
	customBranchColors map[string]int
}

func NewService(configService *config.Service, workingFolder string) *ViewRepo {
	return &ViewRepo{
		RepoChanges:     make(chan RepoChange),
		showRequests:    make(chan ShowRequest),
		currentBranches: make(chan []string),
		branchesGraph:   newBranchesGraph(),
		gitRepo:         gitrepo.NewGitRepo(workingFolder),
		configService:   configService,
		//	customBranchColors: make(map[string]int),
	}
}

func ToSid(commitID string) string {
	return gitrepo.ToSid(commitID)
}

func (t *ViewRepo) StartMonitor(ctx context.Context) {
	go t.monitorViewModelRoutine(ctx)
}

func (t *ViewRepo) TriggerRefreshModel() {
	log.Event("vms-refresh")
	t.gitRepo.TriggerManualRefresh()
}

func (t *ViewRepo) GetCommitDiff(id string) (git.CommitDiff, error) {
	return t.gitRepo.GetCommitDiff(id)
}

func (t *ViewRepo) SwitchToBranch(name string, displayName string, repo Repo) error {
	if strings.HasPrefix(name, "origin/") {
		name = name[7:]
	}
	t.ShowBranch(name, repo)
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

func (t *ViewRepo) Commit(Commit string) error {
	return t.gitRepo.Commit(Commit)
}

func (t *ViewRepo) showBranches(branchIds []string) {
	log.Event("vms-load-repo")
	select {
	case t.showRequests <- ShowRequest{branches: branchIds}:
	default:
	}
}

func (t *ViewRepo) monitorViewModelRoutine(ctx context.Context) {
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
				case t.RepoChanges <- RepoChange{IsStarting: true}:
				case <-ctx.Done():
					return
				}
				break
			}

			log.Infof("Got repo change")

			if change.Error != nil {
				log.Event("vms-error-repo")
				select {
				case t.RepoChanges <- RepoChange{Error: change.Error}:
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

func (t *ViewRepo) triggerSearchRepo(ctx context.Context, searchText string, repo gitrepo.Repo) {
	log.Infof("triggerSearchRepo")
	go func() {
		vRepo := t.getSearchModel(repo, searchText)
		repoChange := RepoChange{SearchText: searchText, ViewRepo: newViewRepo(vRepo)}
		select {
		case t.RepoChanges <- repoChange:
		case <-ctx.Done():
		}
	}()
}

func (t *ViewRepo) triggerFreshViewRepo(ctx context.Context, repo gitrepo.Repo, branchNames []string) {
	log.Infof("triggerFreshViewRepo")
	go func() {
		vRepo := t.getViewModel(repo, branchNames)
		repoChange := RepoChange{ViewRepo: newViewRepo(vRepo)}
		select {
		case t.RepoChanges <- repoChange:
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

func (t *ViewRepo) storeShownBranchesInConfig(branchNames []string) {
	t.configService.SetRepo(t.gitRepo.RepoPath(), func(r *config.Repo) {
		r.ShownBranches = branchNames
	})
}

func (t *ViewRepo) getSearchModel(grepo gitrepo.Repo, searchText string) *viewRepo {
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

func (t *ViewRepo) getViewModel(grepo gitrepo.Repo, branchNames []string) *viewRepo {
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

	// Draw branch lines
	t.branchesGraph.drawBranchLines(repo)

	// Draw branch connector lines
	t.branchesGraph.drawConnectorLines(repo)
	log.Infof("getViewModel done, %s", ti)
	return repo
}

func (t *ViewRepo) adjustCurrentBranchIfStatus(repo *viewRepo) {
	if len(repo.Commits) < 2 || repo.Commits[0].ID != UncommittedID || repo.CurrentCommit == nil {
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

func (t *ViewRepo) BranchColor(name string) cui.Color {
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

// func (s *Service) CurrentBranchNames() []string {
// 	s.lock.Lock()
// 	defer s.lock.Unlock()
// 	if s.currentViewModel == nil {
// 		return []string{}
// 	}
//
// 	var branchNames []string
// 	for _, b := range s.currentViewModel.Branches {
// 		branchNames = append(branchNames, b.name)
// 	}
// 	return branchNames
// }
//
// func (s *Service) GetCommitByIndex(index int) (Commit, error) {
// 	s.lock.Lock()
// 	defer s.lock.Unlock()
// 	if index < 0 || index >= len(s.currentViewModel.Commits) {
// 		return Commit{}, fmt.Errorf("no commit")
// 	}
// 	return toCommit(s.currentViewModel.Commits[index]), nil
// }
//
func (t *ViewRepo) CurrentNotShownBranch(viewRepo Repo) (Branch, bool) {
	current, ok := viewRepo.viewRepo.gitRepo.CurrentBranch()
	if !ok {
		return Branch{}, false
	}
	if containsBranch(viewRepo.viewRepo.Branches, current.Name) {
		return Branch{}, false
	}

	return toBranch(viewRepo.viewRepo.toBranch(current, 0)), true
}

func (t *ViewRepo) CurrentBranch(viewRepo Repo) (Branch, bool) {
	current, ok := viewRepo.viewRepo.gitRepo.CurrentBranch()
	if !ok {
		return Branch{}, false
	}

	for _, b := range viewRepo.viewRepo.Branches {
		if current.Name == b.name {
			return toBranch(b), true
		}
	}

	return toBranch(viewRepo.viewRepo.toBranch(current, 0)), true
}

func (t *ViewRepo) GetAllBranches(viewRepo Repo, skipShown bool) []Branch {
	branches := t.getAllBranches(viewRepo, skipShown)
	sort.SliceStable(branches, func(i, j int) bool {
		return -1 == strings.Compare(branches[i].DisplayName, branches[j].DisplayName)
	})

	return branches
}

func (t *ViewRepo) GetLatestBranches(viewRepo Repo, skipShown bool) []Branch {
	branches := t.getAllBranches(viewRepo, skipShown)
	sort.SliceStable(branches, func(i, j int) bool {
		return viewRepo.viewRepo.gitRepo.CommitByID(branches[i].TipID).AuthorTime.After(viewRepo.viewRepo.gitRepo.CommitByID(branches[j].TipID).AuthorTime)
	})

	return branches
}

func (t *ViewRepo) getAllBranches(viewRepo Repo, skipShown bool) []Branch {
	var branches []Branch
	for _, b := range viewRepo.viewRepo.gitRepo.Branches {
		if containsDisplayNameBranch(branches, b.DisplayName) {
			continue
		}
		if skipShown && containsBranch(viewRepo.viewRepo.Branches, b.Name) {
			continue
		}
		branches = append(branches, toBranch(viewRepo.viewRepo.toBranch(b, 0)))
	}

	return branches
}

func containsDisplayNameBranch(branches []Branch, displayName string) bool {
	for _, b := range branches {
		if displayName == b.DisplayName {
			return true
		}
	}
	return false
}

func (t *ViewRepo) GetCommitOpenInBranches(commitID string, viewRepo Repo) []Branch {
	c := viewRepo.viewRepo.gitRepo.CommitByID(commitID)
	var branches []*gitrepo.Branch

	if len(c.ParentIDs) > 1 {
		// commit has branch merged into this commit add it
		mergeParent := viewRepo.viewRepo.gitRepo.CommitByID(c.ParentIDs[1])
		branches = append(branches, mergeParent.Branch)
	}

	var bs []Branch
	for _, b := range branches {
		if nil != funk.Find(bs, func(bsb Branch) bool {
			return b.Name == bsb.Name || b.RemoteName == bsb.Name ||
				b.Name == bsb.RemoteName
		}) {
			// Skip duplicates
			continue
		}
		if nil != funk.Find(viewRepo.viewRepo.Branches, func(bsb *branch) bool {
			return b.Name == bsb.name
		}) {
			// Skip branches already shown
			continue
		}

		bs = append(bs, toBranch(viewRepo.viewRepo.toBranch(b, 0)))
	}

	return bs
}

func (t *ViewRepo) GetCommitOpenOutBranches(commitID string, viewRepo Repo) []Branch {
	c := viewRepo.viewRepo.gitRepo.CommitByID(commitID)
	var branches []*gitrepo.Branch

	for _, ccId := range c.ChildIDs {
		// commit has children (i.e.), other branches have merged from this branch
		if ccId == UncommittedID {
			continue
		}
		cc := viewRepo.viewRepo.gitRepo.CommitByID(ccId)
		if cc.Branch.Name != c.Branch.Name {
			branches = append(branches, cc.Branch)
		}
	}

	for _, b := range viewRepo.viewRepo.gitRepo.Branches {
		if b.TipID == b.BottomID && b.BottomID == c.Id && b.ParentBranch.Name == c.Branch.Name {
			// empty branch with no own branch commit, (branch start)
			branches = append(branches, b)
		}
	}

	var bs []Branch
	for _, b := range branches {
		if nil != funk.Find(bs, func(bsb Branch) bool {
			return b.Name == bsb.Name || b.RemoteName == bsb.Name ||
				b.Name == bsb.RemoteName
		}) {
			// Skip duplicates
			continue
		}
		if nil != funk.Find(viewRepo.viewRepo.Branches, func(bsb *branch) bool {
			return b.Name == bsb.name
		}) {
			// Skip branches already shown
			continue
		}

		bs = append(bs, toBranch(viewRepo.viewRepo.toBranch(b, 0)))
	}

	return bs
}

//
func (t *ViewRepo) GetShownBranches(viewRepo Repo, skipMaster bool) []Branch {
	var bs []Branch
	for _, b := range viewRepo.viewRepo.Branches {
		if skipMaster && (b.name == masterName || b.name == remoteMasterName) {
			// Do not support closing master branch
			continue
		}
		if b.isRemote && nil != funk.Find(viewRepo.viewRepo.Branches, func(bsb *branch) bool {
			return b.name == bsb.remoteName
		}) {
			// Skip remote if local exist
			continue
		}
		if nil != funk.Find(bs, func(bsb Branch) bool {
			return b.displayName == bsb.DisplayName
		}) {
			// Skip duplicates
			continue
		}

		bs = append(bs, toBranch(b))
	}
	return bs
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

func (t *ViewRepo) ShowBranch(name string, viewRepo Repo) {
	branchNames := t.toBranchNames(viewRepo.viewRepo.Branches)

	branch, ok := viewRepo.viewRepo.gitRepo.BranchByName(name)
	if !ok {
		return
	}

	branchNames = t.addBranchWithAncestors(branchNames, branch)

	log.Event("vms-branches-open")
	t.showBranches(branchNames)
}

func (t *ViewRepo) TriggerSearch(text string) {
	log.Eventf("vms-search", text)
	select {
	case t.showRequests <- ShowRequest{searchText: text}:
	default:
	}
}

//
func (t *ViewRepo) HideBranch(viewRepo Repo, name string) {
	hideBranch, ok := funk.Find(viewRepo.viewRepo.Branches, func(b *branch) bool {
		return name == b.name
	}).(*branch)
	if !ok || hideBranch == nil {
		// No branch with that name
		return
	}
	if hideBranch.remoteName != "" {
		remoteBranch, ok := funk.Find(viewRepo.viewRepo.Branches, func(b *branch) bool {
			return hideBranch.remoteName == b.name
		}).(*branch)
		if ok && remoteBranch != nil {
			// The branch to hide has a remote branch, hiding that and the local branch
			// will be hidden as well
			hideBranch = remoteBranch
		}
	}

	var branchNames []string
	for _, b := range viewRepo.viewRepo.Branches {
		if b.name != hideBranch.name && !hideBranch.isAncestor(b) && b.remoteName != hideBranch.name {
			branchNames = append(branchNames, b.name)
		}
	}

	log.Event("vms-branches-close")
	t.showBranches(branchNames)
}

func (t *ViewRepo) setBranchParentChildRelations(repo *viewRepo) {
	for _, b := range repo.Branches {
		b.tip = repo.commitById[b.tipId]
		b.bottom = repo.commitById[b.bottomId]
		if b.parentBranchName != "" {
			b.parentBranch = repo.BranchByName(b.parentBranchName)
		}
	}
}

func (t *ViewRepo) setParentChildRelations(repo *viewRepo) {
	for _, c := range repo.Commits {
		if len(c.ParentIDs) > 0 {
			// Commit has a parent
			c.Parent = repo.commitById[c.ParentIDs[0]]
			if len(c.ParentIDs) > 1 {
				// Merge parent can be nil if not the merge parent branch is included in the repo view model as well
				c.MergeParent = repo.commitById[c.ParentIDs[1]]
				if c.MergeParent == nil {
					// commit has a merge parent, that is not visible, mark the commit as expandable
					c.More.Set(MoreMergeIn)
				}
			}
		}

		for _, childID := range c.ChildIDs {
			if _, ok := repo.commitById[childID]; !ok {
				// commit has a child, that is not visible, mark the commit as expandable
				c.More.Set(MoreBranchOut)
			}
		}

		if c.More == MoreNone {
			for _, branchName := range c.BranchTips {
				if !repo.containsBranchName(branchName) {
					// Some not shown branch tip att this commit
					c.More.Set(MoreBranchOut)
					break
				}
			}
		}
	}
}

func (t *ViewRepo) toBranchNames(branches []*branch) []string {
	var ids []string
	for _, b := range branches {
		ids = append(ids, b.name)
	}
	return ids
}

func (t *ViewRepo) setAheadBehind(repo *viewRepo) {
	for _, b := range repo.Branches {
		if b.isRemote && b.localName != "" {
			t.setIsRemoteOnlyCommits(repo, b)
		} else if !b.isRemote && b.remoteName != "" {
			t.setLocalOnlyCommits(repo, b)
		}
	}
}
func (t *ViewRepo) setIsRemoteOnlyCommits(repo *viewRepo, b *branch) {
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

func (t *ViewRepo) setLocalOnlyCommits(repo *viewRepo, b *branch) {
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

func (t *ViewRepo) tryGetRealLocalTip(localBranch *branch) *commit {
	var localTip *commit
	if localBranch != nil {
		localTip = localBranch.tip
		if localTip.ID == git.UncommittedID {
			localTip = localTip.Parent
		}
	}
	return localTip
}

func (t *ViewRepo) isSynced(b *branch) bool {
	return b.isRemote && b.localName == "" || !b.isRemote && b.remoteName == ""
}

func (t *ViewRepo) PushBranch(name string) error {
	return t.gitRepo.PushBranch(name)
}

func (t *ViewRepo) PullBranch() error {
	return t.gitRepo.PullBranch()
}

func (t *ViewRepo) getBranchNames(repo *viewRepo) []string {
	var names []string
	for _, b := range repo.Branches {
		names = append(names, b.name)
	}
	return names
}

func (t *ViewRepo) MergeBranch(name string) error {
	return t.gitRepo.MergeBranch(name)
}

func (t *ViewRepo) CreateBranch(name string) error {
	return t.gitRepo.CreateBranch(name)
}

func (t *ViewRepo) DeleteBranch(name string, repo Repo) error {
	branch, ok := repo.viewRepo.gitRepo.BranchByName(name)
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
			if b, ok := repo.viewRepo.gitRepo.BranchByName(branch.LocalName); ok {
				localBranch = b
			}
		}
	}

	if !branch.IsRemote {
		// Local branch, check if there is a corresponding remote branch
		localBranch = branch
		if branch.RemoteName != "" {
			if b, ok := repo.viewRepo.gitRepo.BranchByName(branch.RemoteName); ok {
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

func (t *ViewRepo) addTags(repo *viewRepo, tags []gitrepo.Tag) {
	for _, tag := range tags {
		c, ok := repo.commitById[tag.CommitID]
		if !ok {
			continue
		}
		c.Tags = append(c.Tags, tag.TagName)
	}
}

//
// func (s *Service) ChangeBranchColor(index int) {
// 	log.Event("vms-branches-change-color")
// 	var branchName string
// 	s.lock.Lock()
// 	if index >= len(s.currentViewModel.Commits) {
// 		// Repo must just have changed, just ignore
// 		s.lock.Unlock()
// 		return
// 	}
// 	c := s.currentViewModel.Commits[index]
// 	branchName = c.Branch.displayName
// 	s.lock.Unlock()
//
// 	color := s.BranchColor(branchName)
// 	for i, c := range branchColors {
// 		if color == c {
// 			index := int((i + 1) % len(branchColors))
// 			color = branchColors[index]
// 			s.customBranchColors[branchName] = int(color)
// 			break
// 		}
// 	}
//
// 	s.configService.SetRepo(s.gmRepo.RepoPath, func(r *config.Repo) {
// 		isSet := false
// 		cb := config.Branch{DisplayName: branchName, Color: int(color)}
//
// 		for i, b := range r.Branches {
// 			if branchName == b.DisplayName {
// 				r.Branches[i] = cb
// 				isSet = true
// 				break
// 			}
// 		}
// 		if !isSet {
// 			r.Branches = append(r.Branches, cb)
// 		}
// 	})
// }
//
