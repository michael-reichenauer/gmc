package viewmodel

import (
	"context"
	"github.com/michael-reichenauer/gmc/common/config"
	"github.com/michael-reichenauer/gmc/repoview/viewmodel/gitrepo"
	"github.com/michael-reichenauer/gmc/utils/git"
	"github.com/michael-reichenauer/gmc/utils/log"
	"github.com/michael-reichenauer/gmc/utils/ui"
	"github.com/thoas/go-funk"
	"hash/fnv"
	"strings"
)

const (
	masterName       = "master"
	remoteMasterName = "origin/master"
)

var branchColors = []ui.Color{
	ui.CRed,
	ui.CBlue,
	ui.CYellow,
	ui.CGreen,
	ui.CCyan,
	ui.CRedDk,
	ui.CGreenDk,
	ui.CYellowDk,
	//ui.CBlueDk,
	ui.CMagenta,
	ui.CMagentaDk,
	ui.CCyanDk,
}

type Status struct {
	AllChanges int
	GraphWidth int
}

type Service struct {
	ViewRepos chan ViewRepo

	gitRepo       *gitrepo.GitRepo
	configService *config.Service
	branchesGraph *branchesGraph

	showRequests chan []string

	customBranchColors map[string]int
}

func NewService(configService *config.Service, workingFolder string) *Service {
	return &Service{
		ViewRepos:     make(chan ViewRepo),
		showRequests:  make(chan []string),
		branchesGraph: newBranchesGraph(),
		gitRepo:       gitrepo.NewGitRepo(workingFolder),
		configService: configService,
		//	customBranchColors: make(map[string]int),
	}
}

func ToSid(commitID string) string {
	return gitrepo.ToSid(commitID)
}

func (s *Service) StartMonitor(ctx context.Context) {
	go s.monitorViewModelRoutine(ctx)
}

func (s *Service) TriggerRefreshModel() {
	log.Event("vms-refresh")
	s.gitRepo.TriggerRefreshRepo()
}

func (s *Service) GetCommitDiff(id string) (git.CommitDiff, error) {
	return s.gitRepo.GetCommitDiff(id)
}

func (s *Service) SwitchToBranch(name string) {
	if strings.HasPrefix(name, "origin/") {
		name = name[7:]
	}
	s.gitRepo.SwitchToBranch(name)
}

func (s *Service) showBranches(branchIds []string) {
	log.Event("vms-load-repo")
	select {
	case s.showRequests <- branchIds:
	default:
	}
}

func (s *Service) monitorViewModelRoutine(ctx context.Context) {
	log.Infof("monitorViewModelRoutine start")
	defer log.Infof("View model monitor done")
	s.gitRepo.StartMonitor(ctx)

	var repo gitrepo.Repo
	var branchNames []string

	for {
		select {
		case change, ok := <-s.gitRepo.RepoChanges:
			if !ok {
				return
			}
			log.Infof("Got repo change")
			log.Event("vms-changed-repo")
			if change.Error != nil {
				log.Warnf("Repo error %v", change.Error)
				continue
			}
			repo = change.Repo
			s.triggerFreshViewRepo(ctx, repo, branchNames)
		case names := <-s.showRequests:
			if names == nil {
				// A manual refresh, trigger new repo
				log.Infof("Manual refresh")
				s.gitRepo.TriggerManualRefresh()
				continue
			}
			log.Infof("Refresh of %v", names)
			branchNames = names
			s.triggerFreshViewRepo(ctx, repo, branchNames)
		case <-ctx.Done():
			return
		}
	}
}

func (s *Service) triggerFreshViewRepo(ctx context.Context, repo gitrepo.Repo, branchNames []string) {
	log.Infof("triggerFreshViewRepo")
	go func() {
		vRepo := s.getViewModel(repo, branchNames)
		select {
		case s.ViewRepos <- newViewRepo(vRepo):
			s.configService.SetRepo(s.gitRepo.RepoPath(), func(r *config.Repo) {
				r.ShownBranches = branchNames
			})
		case <-ctx.Done():
		}
	}()
}

func (s *Service) getViewModel(grepo gitrepo.Repo, branchNames []string) *viewRepo {
	log.Infof("getViewModel")
	repo := newRepo()
	repo.gitRepo = grepo
	repo.WorkingFolder = grepo.RepoPath
	repo.UncommittedChanges = grepo.Status.AllChanges()

	branches := s.getGitModelBranches(branchNames, grepo)
	for _, b := range branches {
		repo.addBranch(b)
	}
	currentBranch, ok := grepo.CurrentBranch()
	if ok {
		repo.CurrentBranchName = currentBranch.Name
	}

	repo.addVirtualStatusCommit(grepo)
	for _, c := range grepo.Commits {
		repo.addGitCommit(c)
	}
	s.adjustCurrentBranchIfStatus(repo)
	s.setBranchParentChildRelations(repo)
	s.setParentChildRelations(repo)

	// Draw branch lines
	s.branchesGraph.drawBranchLines(repo)

	// Draw branch connector lines
	s.branchesGraph.drawConnectorLines(repo)
	log.Infof("getViewModel done")
	return repo
}

func (s *Service) adjustCurrentBranchIfStatus(repo *viewRepo) {
	if len(repo.Commits) < 2 || repo.Commits[0].ID != StatusID || repo.CurrentCommit == nil {
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

func (s *Service) BranchColor(name string) ui.Color {
	if strings.HasPrefix(name, "multi:") {
		return ui.CWhite
	}
	color, ok := s.customBranchColors[name]
	if ok {
		return ui.Color(color)
	}
	if name == "master" {
		return ui.CMagenta
	}
	if name == "develop" {
		return ui.CRedDk
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
func (s *Service) CurrentBranch(viewRepo ViewRepo) (Branch, bool) {

	current, ok := viewRepo.viewRepo.gitRepo.CurrentBranch()
	if !ok {
		return Branch{}, false
	}
	if containsBranch(viewRepo.viewRepo.Branches, current.Name) {
		return Branch{}, false
	}

	return toBranch(viewRepo.viewRepo.toBranch(current, 0)), true
}

// func (s *Service) GetAllBranches() []Branch {
// 	s.lock.Lock()
// 	defer s.lock.Unlock()
// 	var branches []Branch
// 	for _, b := range s.gmRepo.Branches {
// 		if containsDisplayNameBranch(branches, b.DisplayName) {
// 			continue
// 		}
// 		if containsBranch(s.currentViewModel.Branches, b.Name) {
// 			continue
// 		}
// 		branches = append(branches, toBranch(s.currentViewModel.toBranch(b, 0)))
// 	}
// 	sort.SliceStable(branches, func(i, j int) bool {
// 		return -1 == strings.Compare(branches[i].DisplayName, branches[j].DisplayName)
// 	})
//
// 	return branches
// }
//
// func (s *Service) GetActiveBranches() []Branch {
// 	s.lock.Lock()
// 	defer s.lock.Unlock()
// 	var branches []Branch
// 	for _, b := range s.gmRepo.Branches {
// 		if containsDisplayNameBranch(branches, b.DisplayName) {
// 			continue
// 		}
// 		if containsBranch(s.currentViewModel.Branches, b.Name) {
// 			continue
// 		}
// 		branches = append(branches, toBranch(s.currentViewModel.toBranch(b, 0)))
// 	}
// 	sort.SliceStable(branches, func(i, j int) bool {
// 		return s.gmRepo.CommitById[branches[i].TipID].AuthorTime.After(s.gmRepo.CommitById[branches[j].TipID].AuthorTime)
// 	})
//
// 	return branches
// }
//
// func containsDisplayNameBranch(branches []Branch, displayName string) bool {
// 	for _, b := range branches {
// 		if displayName == b.DisplayName {
// 			return true
// 		}
// 	}
// 	return false
// }
//
func (s *Service) GetCommitOpenBranches(commitIndex int, viewRepo ViewRepo) []Branch {
	c := viewRepo.viewRepo.Commits[commitIndex]
	if !c.IsMore {
		return nil
	}
	var branches []*gitrepo.Branch

	if len(c.ParentIDs) > 1 {
		// commit has branch merged into this commit add it
		mergeParent := viewRepo.viewRepo.gitRepo.CommitById[c.ParentIDs[1]]
		branches = append(branches, mergeParent.Branch)
	}

	for _, ccId := range c.ChildIDs {
		// commit has children (i.e.), other branches have merged from this branch
		if ccId == StatusID {
			continue
		}
		cc := viewRepo.viewRepo.gitRepo.CommitById[ccId]
		if cc.Branch.Name != c.Branch.name {
			branches = append(branches, cc.Branch)
		}
	}

	for _, b := range viewRepo.viewRepo.gitRepo.Branches {
		if b.TipID == b.BottomID && b.BottomID == c.ID && b.ParentBranch.Name == c.Branch.name {
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
// func (s *Service) GetShownBranches(skipMaster bool) []Branch {
// 	s.lock.Lock()
// 	defer s.lock.Unlock()
//
// 	var bs []Branch
// 	for _, b := range s.currentViewModel.Branches {
// 		if skipMaster && (b.name == masterName || b.name == remoteMasterName) {
// 			// Do not support closing master branch
// 			continue
// 		}
// 		if b.isRemote && nil != funk.Find(s.currentViewModel.Branches, func(bsb *branch) bool {
// 			return b.name == bsb.remoteName
// 		}) {
// 			// Skip remote if local exist
// 			continue
// 		}
// 		if nil != funk.Find(bs, func(bsb Branch) bool {
// 			return b.displayName == bsb.DisplayName
// 		}) {
// 			// Skip duplicates
// 			continue
// 		}
//
// 		bs = append(bs, toBranch(b))
// 	}
// 	return bs
// }
//
func containsBranch(branches []*branch, name string) bool {
	for _, b := range branches {
		if name == b.name {
			return true
		}
	}
	return false
}

func (s *Service) ShowBranch(name string, viewRepo ViewRepo) {
	branchNames := s.toBranchNames(viewRepo.viewRepo.Branches)

	branch, ok := viewRepo.viewRepo.gitRepo.BranchByName(name)
	if !ok {
		return
	}

	branchNames = s.addBranchWithAncestors(branchNames, branch)

	log.Event("vms-branches-open")
	s.showBranches(branchNames)
}

//
// func (s *Service) HideBranch(name string) {
// 	s.lock.Lock()
//
// 	hideBranch, ok := funk.Find(s.currentViewModel.Branches, func(b *branch) bool {
// 		return name == b.name
// 	}).(*branch)
// 	if !ok || hideBranch == nil {
// 		// No branch with that name
// 		s.lock.Unlock()
// 		return
// 	}
// 	if hideBranch.remoteName != "" {
// 		remoteBranch, ok := funk.Find(s.currentViewModel.Branches, func(b *branch) bool {
// 			return hideBranch.remoteName == b.name
// 		}).(*branch)
// 		if ok && remoteBranch != nil {
// 			// The branch to hide has a remote branch, hiding that and the local branch
// 			// will be hidden as well
// 			hideBranch = remoteBranch
// 		}
// 	}
//
// 	var branchNames []string
// 	for _, b := range s.currentViewModel.Branches {
// 		if b.name != hideBranch.name && !hideBranch.isAncestor(b) && b.remoteName != hideBranch.name {
// 			branchNames = append(branchNames, b.name)
// 		}
// 	}
//
// 	s.lock.Unlock()
//
// 	log.Event("vms-branches-close")
// 	s.showBranches(branchNames)
// }
//
// func (s *Service) RepoPath() string {
// 	s.lock.Lock()
// 	defer s.lock.Unlock()
// 	return s.gmRepo.RepoPath
// }
//
// func (s *Service) CloseBranch(index int) {
// 	s.lock.Lock()
// 	if index >= len(s.currentViewModel.Commits) {
// 		// Repo must just have changed, just ignore
// 		s.lock.Unlock()
// 		return
// 	}
// 	c := s.currentViewModel.Commits[index]
// 	if c.Branch.name == masterName || c.Branch.name == remoteMasterName {
// 		// Cannot close master
// 		s.lock.Unlock()
// 		return
// 	}
//
// 	// get branch ids except for the commit branch or decedent branches
// 	var branchNames []string
// 	for _, b := range s.currentViewModel.Branches {
// 		if b.name != c.Branch.name && !c.Branch.isAncestor(b) {
// 			branchNames = append(branchNames, b.name)
// 		}
// 	}
// 	s.lock.Unlock()
// 	log.Event("vms-branches-close")
// 	s.showBranches(branchNames)
// }

func (s *Service) setBranchParentChildRelations(repo *viewRepo) {
	for _, b := range repo.Branches {
		b.tip = repo.commitById[b.tipId]
		b.bottom = repo.commitById[b.bottomId]
		if b.parentBranchName != "" {
			b.parentBranch = repo.BranchByName(b.parentBranchName)
		}
	}
}

func (s *Service) setParentChildRelations(repo *viewRepo) {
	for _, c := range repo.Commits {
		if len(c.ParentIDs) > 0 {
			// Commit has a parent
			c.Parent = repo.commitById[c.ParentIDs[0]]
			if len(c.ParentIDs) > 1 {
				// Merge parent can be nil if not the merge parent branch is included in the repo view model as well
				c.MergeParent = repo.commitById[c.ParentIDs[1]]
				if c.MergeParent == nil {
					// commit has a merge parent, that is not visible, mark the commit as expandable
					c.IsMore = true
				}
			}
		}

		for _, childID := range c.ChildIDs {
			if _, ok := repo.commitById[childID]; !ok {
				// commit has a child, that is not visible, mark the commit as expandable
				c.IsMore = true
			}
		}

		if !c.IsMore {
			for _, branchName := range c.BranchTips {
				if !repo.containsBranchName(branchName) {
					c.IsMore = true
					break
				}
			}

		}
	}
}

func (s *Service) toBranchNames(branches []*branch) []string {
	var ids []string
	for _, b := range branches {
		ids = append(ids, b.name)
	}
	return ids
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
