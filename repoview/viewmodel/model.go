package viewmodel

import (
	"fmt"
	"github.com/michael-reichenauer/gmc/repoview/viewmodel/gitmodel"
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

type Model struct {
	ChangedEvents chan interface{}
	gitModel      *gitmodel.Handler

	lock         sync.Mutex
	currentRepox *repo
	gmRepo       gitmodel.Repo
	gmStatus     gitmodel.Status
	err          error
}

func NewModel(repoPath string) *Model {
	gm := gitmodel.NewModel(repoPath)
	return &Model{
		ChangedEvents: make(chan interface{}),
		gitModel:      gm,
		currentRepox:  newRepo(),
	}
}

func (h *Model) Start() {
	h.gitModel.Start()
	go h.monitorGitModelRoutine()
}

func (h *Model) monitorGitModelRoutine() {
	for {
		select {
		case gmRepo := <-h.gitModel.RepoEvents:
			h.lock.Lock()
			h.gmRepo = gmRepo
			h.gmStatus = gmRepo.Status
			h.lock.Unlock()

		case gmStatus := <-h.gitModel.StatusEvents:
			h.lock.Lock()
			h.gmStatus = gmStatus
			h.lock.Unlock()
		}
		var branchIds []string
		h.lock.Lock()
		for _, b := range h.currentRepox.Branches {
			branchIds = append(branchIds, b.name)
		}
		h.lock.Unlock()
		h.loadBranches(branchIds)
	}
}

func (h *Model) TriggerRefresh() {
	h.gitModel.TriggerRefresh()
}

//func (h *Model) Load() {
//	t := time.Now()
//	h.gitModel.Load()
//	gmRepo := h.gitModel.GetRepo()
//	gmStatus := h.gitModel.GetStatus()
//	h.LoadBranches([]string{}, gmRepo, gmStatus)
//	log.Infof("Load time %v", time.Since(t))
//}

func (h *Model) loadBranches(branchIds []string) {
	t := time.Now()
	repo := h.getRepoModel(branchIds)
	log.Infof("LoadBranches time %v", time.Since(t))
	h.lock.Lock()
	h.currentRepox = repo
	h.lock.Unlock()
	h.ChangedEvents <- nil
}

func (h *Model) GetCommitByIndex(index int) (Commit, error) {
	if h.err != nil {
		return Commit{}, h.err
	}
	h.lock.Lock()
	defer h.lock.Unlock()
	if index < 0 || index >= len(h.currentRepox.Commits) {
		return Commit{}, fmt.Errorf("no commit")
	}
	return toCommit(h.currentRepox.Commits[index]), nil
}

func (h *Model) GetRepoViewPort(firstIndex, count int) (ViewPort, error) {
	if h.err != nil {
		return ViewPort{}, h.err
	}

	h.lock.Lock()
	defer h.lock.Unlock()

	if count > len(h.currentRepox.Commits) {
		// Requested count larger than available, return just all available commits
		count = len(h.currentRepox.Commits)
	}

	if firstIndex+count >= len(h.currentRepox.Commits) {
		// Requested commits past available, adjust to return available commits
		firstIndex = len(h.currentRepox.Commits) - count
	}

	return newViewPort(h.currentRepox, firstIndex, count), nil
}

func (h *Model) OpenBranch(index int) {
	h.lock.Lock()
	if index >= len(h.currentRepox.Commits) {
		// Repo must just have changed, just ignore
		return
	}
	c := h.currentRepox.Commits[index]
	if !c.IsMore {
		// Not a point that can be expanded
		return
	}

	branchIds := h.toBranchIds(h.currentRepox.Branches)

	if len(c.ParentIDs) > 1 {
		// commit has branch merged into this commit add it (if not already added
		mergeParent := h.gmRepo.CommitById[c.ParentIDs[1]]
		branchIds = h.addBranchWithAncestors(branchIds, mergeParent.Branch)
	}
	for _, ccId := range c.ChildIDs {
		cc := h.gmRepo.CommitById[ccId]
		if cc.Branch.Name != c.Branch.name {
			branchIds = h.addBranchWithAncestors(branchIds, cc.Branch)
		}
	}
	for _, b := range h.gmRepo.Branches {
		if b.TipID == b.BottomID && b.BottomID == c.ID && b.ParentBranch.Name == c.Branch.name {
			// empty branch with no own branch commit, (branch start)
			branchIds = h.addBranchWithAncestors(branchIds, b)
		}
	}
	for _, b := range h.gmRepo.Branches {
		i1 := utils.StringsIndex(branchIds, b.RemoteName)
		i2 := utils.StringsIndex(branchIds, b.Name)
		if i2 == -1 && i1 != -1 {
			// a remote branch is included, but not its local branch
			branchIds = append(branchIds, "")
			copy(branchIds[i1+1:], branchIds[i1:])
			branchIds[i1] = b.Name
		}
	}
	h.lock.Unlock()
	h.loadBranches(branchIds)
}

func (h *Model) CloseBranch(index int) {
	h.lock.Lock()
	if index >= len(h.currentRepox.Commits) {
		// Repo must just have changed, just ignore
		return
	}
	c := h.currentRepox.Commits[index]
	if c.Branch.name == masterName || c.Branch.name == remoteMasterName {
		// Cannot close master
		return
	}

	// get branch ids except for the commit branch or decedent branches
	var branchIds []string
	for _, b := range h.currentRepox.Branches {
		if b.name != c.Branch.name && !c.Branch.isAncestor(b) {
			branchIds = append(branchIds, b.name)
		}
	}
	h.lock.Unlock()
	h.loadBranches(branchIds)
}

func (h *Model) Refresh(viewPort ViewPort) {
	//t := time.Now()
	//var branchIds []string
	//for _, b := range viewPort.repo.Branches {
	//	branchIds = append(branchIds, b.name)
	//}
	//h.gitModel.Load()
	//
	//h.LoadBranches(branchIds, gmRepo, gmStatus)
	//log.Infof("Refresh time %v", time.Since(t))
}

func (h *Model) getRepoModel(branchIds []string) *repo {
	repo := newRepo()
	repo.gmRepo = h.gmRepo
	repo.gmStatus = h.gmStatus

	branches := h.getGitModelBranches(branchIds, h.gmRepo, h.gmStatus)
	for _, b := range branches {
		repo.addBranch(b)
	}
	currentBranch, ok := h.gmRepo.CurrentBranch()
	if ok {
		repo.CurrentBranchName = currentBranch.Name
	}

	repo.addVirtualStatusCommit()
	for _, c := range repo.gmRepo.Commits {
		repo.addGitCommit(c)
	}

	h.setParentChildRelations(repo)

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
							cc.graph[i+1].Connect.Set(BMLine)
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

func (h *Model) setParentChildRelations(repo *repo) {
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

func (h *Model) toBranchIds(branches []*branch) []string {
	var ids []string
	for _, b := range branches {
		ids = append(ids, b.name)
	}
	return ids
}