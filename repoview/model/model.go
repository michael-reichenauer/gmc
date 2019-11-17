package model

import (
	"github.com/michael-reichenauer/gmc/repoview/model/gitmodel"
	"github.com/michael-reichenauer/gmc/utils/log"
	"strings"
	"sync"
)

const masterID = "master:local"

type Model struct {
	gitModel *gitmodel.Handler

	lock        sync.Mutex
	currentRepo *repo

	err error
}

func NewModel(repoPath string) *Model {
	return &Model{
		gitModel:    gitmodel.NewModel(repoPath),
		currentRepo: newRepo(),
	}
}

func (h *Model) Load() {
	h.gitModel.Load()
	gmRepo := h.gitModel.GetRepo()
	h.LoadBranches([]string{}, gmRepo)
}

func (h *Model) LoadBranches(branchIds []string, gmRepo gitmodel.Repo) {
	repo := h.getRepoModel(branchIds, gmRepo)
	h.lock.Lock()
	h.currentRepo = repo
	h.lock.Unlock()
}

func (h *Model) GetRepoViewPort(first, last int, selected int) (ViewPort, error) {
	if h.err != nil {
		return ViewPort{}, h.err
	}

	h.lock.Lock()
	defer h.lock.Unlock()
	return newViewPort(h.currentRepo, first, last, selected), nil
}

func (h *Model) OpenBranch(viewPort ViewPort, index int) {
	c := viewPort.repo.Commits[index]
	if !c.IsMore {
		// Not a point that can be expanded
		return
	}

	branchIds := h.toBranchIds(viewPort.repo.Branches)

	if len(c.ParentIDs) > 1 {
		// commit has branch merged into this commit add it (if not already added
		mergeParent := viewPort.repo.gitRepo.CommitById(c.ParentIDs[1])
		branchIds = h.addBranch(branchIds, mergeParent.Branch)
	}
	for _, ccId := range c.ChildIDs {
		cc := viewPort.repo.gitRepo.CommitById(ccId)
		if cc.Branch.ID != c.Branch.id {
			branchIds = h.addBranch(branchIds, cc.Branch)
		}
	}

	h.LoadBranches(branchIds, viewPort.repo.gitRepo)
}

func (h *Model) CloseBranch(viewPort ViewPort, index int) {
	c := viewPort.repo.Commits[index]
	if c.Branch.id == "master:local" {
		// Cannot close master
		return
	}

	// get branch ids except for the commit branch or decedent branches
	var branchIds []string
	for _, b := range viewPort.repo.Branches {
		if b.id != c.Branch.id && !c.Branch.isAncestor(b) {
			branchIds = append(branchIds, b.id)
		}
	}

	h.LoadBranches(branchIds, viewPort.repo.gitRepo)
}

func (h *Model) getRepoModel(branchIds []string, gRepo gitmodel.Repo) *repo {
	repo := newRepo()
	repo.gitRepo = gRepo

	if len(branchIds) == 0 {
		currentBranch, ok := repo.gitRepo.CurrentBranch()
		if ok {
			branchIds = h.addBranch(branchIds, currentBranch)
		} else {
			branchIds = h.addBranchId(branchIds, masterID)
		}
	}

	for _, id := range branchIds {
		branch, ok := repo.gitRepo.BranchByID(id)
		if ok {
			repo.addBranch(branch)
		}
	}

	for _, c := range repo.gitRepo.Commits {
		if strings.HasPrefix(c.Id, "81e1") {
			log.Debugf("")
		}
		repo.addCommit(c)
	}

	h.setParentChildRelations(repo)

	// Draw branch lines
	for _, b := range repo.Branches {
		b.tip = repo.commitById[b.tipId]
		c := b.tip
		for {
			if c.Branch != b {
				// this commit is not part of the branch (multiple branched on the same commit)
				if b.tipId == c.ID {
					b.bottom = c
				}
				break
			}
			if c == c.Branch.tip && c.Branch.isGitBranch && c != c.Branch.bottom {
				c.graph[b.index].Branch.Set(BCommit)
			} else if c == c.Branch.tip && c.Branch.isGitBranch {
				c.graph[b.index].Branch.Set(BTip)
			} else if c == c.Branch.tip {
				c.graph[b.index].Branch.Set(BTip)
			} else if c == c.Branch.bottom {
				c.graph[b.index].Branch.Set(BBottom)
			} else {
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
		for i, b := range repo.Branches {
			c.graph[i].BranchId = b.id
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
				} else if c.Parent != nil && c.Parent.Branch != c.Branch {
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
				} else if c.Index >= b.tip.Index && c.Index <= b.bottom.Index {
					c.graph[i].Branch.Set(BLine)
				}
			}

		}
	}
	return repo
}

func (h *Model) setParentChildRelations(repo *repo) {
	for _, c := range repo.Commits {
		if len(c.ParentIDs) > 0 {
			// if a commit has a parent, it is included in the repo model
			c.Parent = repo.commitById[c.ParentIDs[0]]
			if c.Branch != c.Parent.Branch {
				// The parent branch is different, reached a bottom/beginning, i.e. knows the parent branch
				c.Branch.bottom = c
				c.Branch.parentBranch = c.Parent.Branch
				c.Parent.Branch.childBranches = append(c.Parent.Branch.childBranches, c.Branch)
			}
			if len(c.ParentIDs) > 1 {
				// Merge parent can be nil if not the merge parent branch is included in the repo model as well
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
		ids = append(ids, b.id)
	}
	return ids
}

func (h *Model) addBranch(branchIds []string, branch *gitmodel.Branch) []string {
	ids := h.branchAncestorIDs(branch)
	for _, id := range ids {
		branchIds = h.addBranchId(branchIds, id)
	}
	return branchIds
}

func (h *Model) addBranchId(branchIds []string, branchId string) []string {
	isAdded := false
	for _, id := range branchIds {
		if id == branchId {
			isAdded = true
		}
	}
	if !isAdded {
		branchIds = append(branchIds, branchId)
	}
	return branchIds
}

func (h *Model) branchAncestorIDs(b *gitmodel.Branch) []string {
	var ids []string
	for cb := b; cb != nil; cb = cb.ParentBranch {
		ids = append(ids, cb.ID)
	}
	for i := len(ids)/2 - 1; i >= 0; i-- {
		opp := len(ids) - 1 - i
		ids[i], ids[opp] = ids[opp], ids[i]
	}
	return ids
}
