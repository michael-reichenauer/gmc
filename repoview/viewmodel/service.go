package viewmodel

import (
	"fmt"
	"github.com/michael-reichenauer/gmc/repoview/viewmodel/gitrepo"
	"github.com/michael-reichenauer/gmc/utils"
	"github.com/michael-reichenauer/gmc/utils/log"
	"sync"
	"time"
)

const (
	masterName       = "master"
	remoteMasterName = "origin/master"
)

type Status struct {
	AllChanges int
	GraphWidth int
}

type Service struct {
	ChangedEvents  chan interface{}
	gitRepoService *gitrepo.Service

	lock             sync.Mutex
	currentViewModel *repo
	gmRepo           gitrepo.Repo
	gmStatus         gitrepo.Status
}

func NewModel(repoPath string) *Service {
	gm := gitrepo.NewModel(repoPath)
	return &Service{
		ChangedEvents:    make(chan interface{}),
		gitRepoService:   gm,
		currentViewModel: newRepo(),
	}
}

func (s *Service) Start() {
	s.gitRepoService.StartRepoMonitor()
	go s.monitorGitModelRoutine()
}

func (s *Service) TriggerRefreshModel() {
	s.gitRepoService.TriggerRefreshRepo()
}

func (s *Service) LoadRepo(branchIds []string) {
	gmRepo, err := s.gitRepoService.GetFreshRepo()
	if err != nil {
		log.Infof("Detected repo error: %v", err)
		return
	}
	s.lock.Lock()
	s.gmRepo = gmRepo
	s.gmStatus = gmRepo.Status
	s.lock.Unlock()
	repo := s.getViewModel(branchIds)
	s.lock.Lock()
	s.currentViewModel = repo
	s.lock.Unlock()
}

func (s *Service) showBranches(branchIds []string) {
	t := time.Now()
	repo := s.getViewModel(branchIds)
	log.Infof("LoadBranches time %v", time.Since(t))
	s.lock.Lock()
	s.currentViewModel = repo
	s.lock.Unlock()
	s.ChangedEvents <- nil
}

func (s *Service) monitorGitModelRoutine() {
	for {
		select {
		case gmRepo := <-s.gitRepoService.RepoEvents:
			log.Infof("Detected repo change")
			s.lock.Lock()
			s.gmRepo = gmRepo
			s.gmStatus = gmRepo.Status
			s.lock.Unlock()

		case gmStatus := <-s.gitRepoService.StatusEvents:
			log.Infof("Detected status change")
			s.lock.Lock()
			s.gmStatus = gmStatus
			s.lock.Unlock()
		case err := <-s.gitRepoService.ErrorEvents:
			log.Infof("Detected repo error: %v", err)
		}

		// Refresh model
		s.showBranches(s.CurrentBranchNames())
	}
}

func (s *Service) CurrentBranchNames() []string {
	s.lock.Lock()
	defer s.lock.Unlock()
	if s.currentViewModel == nil {
		return []string{}
	}

	var branchNames []string
	for _, b := range s.currentViewModel.Branches {
		branchNames = append(branchNames, b.name)
	}
	return branchNames
}

func (s *Service) GetCommitByIndex(index int) (Commit, error) {
	s.lock.Lock()
	defer s.lock.Unlock()
	if index < 0 || index >= len(s.currentViewModel.Commits) {
		return Commit{}, fmt.Errorf("no commit")
	}
	return toCommit(s.currentViewModel.Commits[index]), nil
}

func (s *Service) GetRepoViewPort(firstIndex, count int) (ViewPort, error) {
	s.lock.Lock()
	defer s.lock.Unlock()

	if count > len(s.currentViewModel.Commits) {
		// Requested count larger than available, return just all available commits
		count = len(s.currentViewModel.Commits)
	}

	if firstIndex+count >= len(s.currentViewModel.Commits) {
		// Requested commits past available, adjust to return available commits
		firstIndex = len(s.currentViewModel.Commits) - count
	}

	return newViewPort(s.currentViewModel, firstIndex, count), nil
}

func (s *Service) OpenBranch(index int) {
	s.lock.Lock()
	if index >= len(s.currentViewModel.Commits) {
		// Repo must just have changed, just ignore
		s.lock.Unlock()
		return
	}
	c := s.currentViewModel.Commits[index]
	if !c.IsMore {
		// Not a point that can be expanded
		s.lock.Unlock()
		return
	}

	branchIds := s.toBranchIds(s.currentViewModel.Branches)

	if len(c.ParentIDs) > 1 {
		// commit has branch merged into this commit add it (if not already added
		mergeParent := s.gmRepo.CommitById[c.ParentIDs[1]]
		branchIds = s.addBranchWithAncestors(branchIds, mergeParent.Branch)
	}
	for _, ccId := range c.ChildIDs {
		cc := s.gmRepo.CommitById[ccId]
		if cc.Branch.Name != c.Branch.name {
			branchIds = s.addBranchWithAncestors(branchIds, cc.Branch)
		}
	}
	for _, b := range s.gmRepo.Branches {
		if b.TipID == b.BottomID && b.BottomID == c.ID && b.ParentBranch.Name == c.Branch.name {
			// empty branch with no own branch commit, (branch start)
			branchIds = s.addBranchWithAncestors(branchIds, b)
		}
	}
	for _, b := range s.gmRepo.Branches {
		i1 := utils.StringsIndex(branchIds, b.RemoteName)
		i2 := utils.StringsIndex(branchIds, b.Name)
		if i2 == -1 && i1 != -1 {
			// a remote branch is included, but not its local branch
			branchIds = append(branchIds, "")
			copy(branchIds[i1+1:], branchIds[i1:])
			branchIds[i1] = b.Name
		}
	}
	s.lock.Unlock()
	s.showBranches(branchIds)
}

func (s *Service) CloseBranch(index int) {
	s.lock.Lock()
	if index >= len(s.currentViewModel.Commits) {
		// Repo must just have changed, just ignore
		s.lock.Unlock()
		return
	}
	c := s.currentViewModel.Commits[index]
	if c.Branch.name == masterName || c.Branch.name == remoteMasterName {
		// Cannot close master
		s.lock.Unlock()
		return
	}

	// get branch ids except for the commit branch or decedent branches
	var branchIds []string
	for _, b := range s.currentViewModel.Branches {
		if b.name != c.Branch.name && !c.Branch.isAncestor(b) {
			branchIds = append(branchIds, b.name)
		}
	}
	s.lock.Unlock()
	s.showBranches(branchIds)
}

func (s *Service) getViewModel(branchIds []string) *repo {
	repo := newRepo()
	repo.gmRepo = s.gmRepo
	repo.gmStatus = s.gmStatus

	branches := s.getGitModelBranches(branchIds, s.gmRepo, s.gmStatus)
	for _, b := range branches {
		repo.addBranch(b)
	}
	currentBranch, ok := s.gmRepo.CurrentBranch()
	if ok {
		repo.CurrentBranchName = currentBranch.Name
	}

	repo.addVirtualStatusCommit()
	for _, c := range repo.gmRepo.Commits {
		repo.addGitCommit(c)
	}

	s.setParentChildRelations(repo)

	// Draw branch lines
	for _, b := range repo.Branches {
		b.tip = repo.commitById[b.tipId]
		c := b.tip
		for {
			if c.Branch != b {
				// this commit is not part of the branch (multiple branched on the same commit)
				break
			}
			if c == c.Branch.tip {
				c.graph[b.index].Branch.Set(BTip)
			}
			if c == c.Branch.tip && c.Branch.isGitBranch {
				c.graph[b.index].Branch.Set(BActiveTip)
			}
			if c == c.Branch.bottom {
				c.graph[b.index].Branch.Set(BBottom)
			}
			if c != c.Branch.tip && c != c.Branch.bottom {
				c.graph[b.index].Branch.Set(BCommit)
			}

			if c.Parent == nil || c.Branch != c.Parent.Branch {
				// Reached bottom of branch
				break
			}
			c = c.Parent
		}
	}

	// Draw branch connector lines
	for _, c := range repo.Commits {
		if c.ID == StatusID {
			continue
		}
		for i, b := range repo.Branches {
			c.graph[i].BranchName = b.name
			c.graph[i].BranchDisplayName = b.displayName
			if c.Branch == b {
				// Commit branch
				if c.MergeParent != nil {
					// Commit has merge (2 parents)
					if c.MergeParent.Branch.index < c.Branch.index {
						// Other branch is left side ╭
						c.graph[i].Connect.Set(BMergeLeft)
						c.graph[i].Branch.Set(BMergeLeft)
						// Draw horizontal pass through line ───
						for k := c.MergeParent.Branch.index + 1; k < c.Branch.index; k++ {
							c.MergeParent.graph[k].Connect.Set(BPass)
							c.MergeParent.graph[k].Branch.Set(BPass)
						}
						// Draw vertical down line │
						for j := c.Index + 1; j < c.MergeParent.Index; j++ {
							cc := repo.Commits[j]
							cc.graph[i].Connect.Set(BMLine)
						}
						c.MergeParent.graph[i].Connect.Set(BBranchRight) //  ╯
					} else {
						// Other branch is right side  ╮
						// Draw merge in rune
						c.graph[c.MergeParent.Branch.index].Connect.Set(BMergeRight)
						// Draw horizontal pass through line ───
						for k := i + 1; k < c.MergeParent.Branch.index; k++ {
							c.graph[k].Connect.Set(BPass)
							c.graph[k].Branch.Set(BPass)
						}
						// Draw vertical down line │
						for j := c.Index + 1; j < c.MergeParent.Index; j++ {
							cc := repo.Commits[j]
							cc.graph[c.MergeParent.Branch.index].Connect.Set(BMLine)
						}
						// Draw branch out rune ╰
						c.MergeParent.graph[c.MergeParent.Branch.index].Connect.Set(BBranchLeft)
						c.MergeParent.graph[c.MergeParent.Branch.index].Branch.Set(BBranchLeft)
					}
				}
				if c.Parent != nil && c.Parent.Branch != c.Branch {
					// Commit parent is on other branch (bottom/first commit on this branch)
					if c.Parent.Branch.index < c.Branch.index {
						// Other branch is left side e
						c.graph[i].Connect.Set(BMergeLeft)
						c.graph[i].Branch.Set(BMergeLeft)
						for k := c.Parent.Branch.index + 1; k < c.Branch.index; k++ {
							c.Parent.graph[k].Connect.Set(BPass)
							c.Parent.graph[k].Branch.Set(BPass)
						}
						for j := c.Index + 1; j < c.Parent.Index; j++ {
							cc := repo.Commits[j]
							cc.graph[i].Connect.Set(BMLine)
						}
						c.Parent.graph[i].Connect.Set(BBranchRight)
					} else {
						// Other branch is right side
						c.graph[i+1].Connect.Set(BMergeRight)
						for j := c.Index + 1; j < c.Parent.Index; j++ {
							cc := repo.Commits[j]
							cc.graph[i+1].Connect.Set(BMLine)
						}
						c.Parent.graph[c.Parent.Branch.index].Connect.Set(BBranchLeft)
						c.Parent.graph[c.Parent.Branch.index].Branch.Set(BBranchLeft)
					}
				}
			} else {
				// Other branch
				if b.tip == c {
					// this branch tip does not have a branch of its own,
					c.graph[i].Branch.Set(BBottom | BPass)
					for k := c.Branch.index + 1; k <= i; k++ {
						c.graph[k].Connect.Set(BPass)
						c.graph[k].Branch.Set(BPass)
					}
				} else if c.Index >= b.tip.Index && c.Index <= b.bottom.Index {
					c.graph[i].Branch.Set(BLine)
				}
			}

		}
	}
	return repo
}

func (s *Service) setParentChildRelations(repo *repo) {
	for _, b := range repo.Branches {
		b.tip = repo.commitById[b.tipId]
		b.bottom = repo.commitById[b.bottomId]
		if b.parentBranchID != "" {
			b.parentBranch = repo.BranchById(b.parentBranchID)
		}
		//if b.tipId == b.bottomId && b.bottom.Branch != b {
		//	// an empty branch, with no own branch commit, (branch start)
		//	b.bottom.IsMore = true
		//}
	}

	for _, c := range repo.Commits {
		if len(c.ParentIDs) > 0 {
			// if a commit has a parent, it is included in the repo viewmodel
			c.Parent = repo.commitById[c.ParentIDs[0]]
			if len(c.ParentIDs) > 1 {
				// Merge parent can be nil if not the merge parent branch is included in the repo viewmodel as well
				c.MergeParent = repo.commitById[c.ParentIDs[1]]
			}
		} else {
			// No parent, reached initial commit of the repo
			c.Branch.bottom = c
		}
		if c.MergeParent == nil && len(c.ParentIDs) > 1 {
			// commit has a merge parent, that is not visible, mark the commit as expandable
			c.IsMore = true
		}
		for _, ccId := range c.ChildIDs {
			if _, ok := repo.commitById[ccId]; !ok {
				// commit has a child, that is not visible, mark the commit as expandable
				c.IsMore = true
			}
		}
	}
}

func (s *Service) toBranchIds(branches []*branch) []string {
	var ids []string
	for _, b := range branches {
		ids = append(ids, b.name)
	}
	return ids
}

func (s *Service) RepoPath() string {
	s.lock.Lock()
	defer s.lock.Unlock()
	return s.gmRepo.RepoPath
}
