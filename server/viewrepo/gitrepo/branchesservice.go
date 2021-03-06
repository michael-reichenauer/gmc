package gitrepo

import (
	"github.com/michael-reichenauer/gmc/utils"
)

// Default branch priority determines parent child branch relations.
var DefaultBranchPriority = []string{"origin/master", "master", "origin/develop", "develop"}

type branchesService struct {
	branchNames *branchNameParser
}

func newBranchesService() *branchesService {
	return &branchesService{branchNames: newBranchNameParser()}
}

func (h *branchesService) setBranchForAllCommits(repo *Repo) {
	h.setGitBranchTips(repo)
	h.setCommitBranchesAndChildren(repo)
	h.determineCommitBranches(repo)
	h.determineBranchHierarchy(repo)
}

func (h *branchesService) setGitBranchTips(repo *Repo) {
	var missingBranches []string
	for _, b := range repo.Branches {
		tip, ok := repo.TryGetCommitByID(b.TipID)
		if !ok {
			missingBranches = append(missingBranches, b.Name)
			continue
		}
		tip.addBranch(b)
		tip.BranchTipNames = append(tip.BranchTipNames, b.Name)
		if b.IsCurrent {
			tip.IsCurrent = true
		}
	}
	if missingBranches != nil {
		var branches []*Branch
		for _, b := range repo.Branches {
			if utils.StringsContains(missingBranches, b.Name) {
				continue
			}
			branches = append(branches, b)
		}
		repo.Branches = branches
	}
}

func (h *branchesService) setCommitBranchesAndChildren(repo *Repo) {
	for _, c := range repo.Commits {
		h.branchNames.parseCommit(c)
		if len(c.ParentIDs) == 2 && h.branchNames.isPullMerge(c) {
			// if the commit is a pull merger, we do switch the order of parents
			c.ParentIDs = []string{c.ParentIDs[1], c.ParentIDs[0]}
		}

		parent, ok := repo.Parent(c, 0)
		if ok {
			parent.Children = append(parent.Children, c)
			parent.ChildIDs = append(parent.ChildIDs, c.Id)
			parent.addBranches(c.Branches)
			c.FirstParent = parent
		}

		mergeParent, ok := repo.Parent(c, 1)
		if ok {
			mergeParent.MergeChildren = append(mergeParent.MergeChildren, c)
			mergeParent.ChildIDs = append(mergeParent.ChildIDs, c.Id)
			c.MergeParent = mergeParent
		}
	}
}

func (h *branchesService) determineBranchHierarchy(repo *Repo) {
	for _, b := range repo.Branches {
		if b.BottomID == "" {
			b.BottomID = b.TipID
		}

		bottom := repo.CommitByID(b.BottomID)
		if bottom.Branch != b {
			// the tip does not own the tip commit, i.e. a branch pointer to another branch
			b.ParentBranch = bottom.Branch
		} else if bottom.FirstParent != nil {
			b.ParentBranch = bottom.FirstParent.Branch
		}
	}
}

func (h *branchesService) determineCommitBranches(repo *Repo) {
	for _, c := range repo.Commits {
		h.determineBranch(repo, c)
		h.setMasterBackbone(c)
		c.Branch.BottomID = c.Id
	}
}

func (h *branchesService) setMasterBackbone(c *Commit) {
	if c.FirstParent == nil {
		return
	}
	if utils.StringsContains(DefaultBranchPriority, c.Branch.Name) {
		// master and develop are special and will make a "backbone" for other branches to depend on
		c.FirstParent.addBranch(c.Branch)
	}
}

func (h *branchesService) determineBranch(repo *Repo, c *Commit) {
	if len(c.Branches) == 1 {
		// Commit only has one branch, use that
		c.Branch = c.Branches[0]
		return
	}

	if branch := h.isLocalRemoteBranch(c); branch != nil {
		// local and remote branch, prefer remote
		c.Branch = branch
		return
	}

	if branch := h.isMergedDeletedBranch(repo, c); branch != nil {
		// Commit has no branch, must be a deleted branch tip merged into some other branch
		c.Branch = branch
		c.addBranch(c.Branch)
		return
	}

	if len(c.Branches) == 0 && len(c.Children) == 1 {
		// commit has no known branches (middle commit in deleted branch), but has one child, use same branch
		c.Branch = c.Children[0].Branch
		c.addBranch(c.Branch)
		return
	}

	if len(c.Children) == 1 && c.Children[0].isLikely {
		// commit has one child, which has a likely known branch, use same branch
		c.Branch = c.Children[0].Branch
		c.isLikely = true
		return
	}

	if branch := h.hasPriorityBranch(c); branch != nil {
		// Commit, has many possible branches, and one is in the priority list, e.g. master, develop, ...
		c.Branch = branch
		return
	}

	if name := h.branchNames.branchName(c.Id); name != "" {
		// The commit branch name could be parsed form the subject (or a child subject).
		// Lets use that as a branch and also let children use that branch if they only are multi branch
		branch := h.tryGetBranchFromName(c, name)
		var current *Commit
		if branch != nil && branch.BottomID != "" {
			current = repo.CommitByID(branch.BottomID)
		}
		if current == nil {
			for current = c; len(current.Children) == 1 && current.Children[0].Branch.IsMultiBranch; current = current.Children[0] {
			}
		}
		if branch != nil {
			for ; current != c.FirstParent; current = current.FirstParent {
				current.Branch = branch
				current.isLikely = true
				current.addBranch(branch)
			}
			c.Branch.BottomID = c.Id
			return
		}
	}

	if branch := h.isChildMultiBranch(c); branch != nil {
		// one of the commit children is a multi branch, reuse
		c.Branch = branch
		c.addBranch(c.Branch)
		return
	}

	// Commit, has several possible branches, create a new multi branch
	c.Branch = repo.addMultiBranch(c)
	c.addBranch(c.Branch)
}

func (h *branchesService) hasPriorityBranch(c *Commit) *Branch {
	if len(c.Branches) < 1 {
		return nil
	}
	for _, bp := range DefaultBranchPriority {
		for _, cb := range c.Branches {
			if bp == cb.Name {
				return cb
			}
		}
	}
	return nil
}

func (h *branchesService) isChildMultiBranch(c *Commit) *Branch {
	for _, cc := range c.Children {
		if cc.Branch != nil && cc.Branch.IsMultiBranch {
			// one of the commit children is a multi branch
			return cc.Branch
		}
	}
	return nil
}

func (h *branchesService) tryGetBranchFromName(c *Commit, name string) *Branch {
	// Try find a branch with the name
	for _, b := range c.Branches {
		if name == b.Name {
			// Found a branch, if the branch has a remote branch, try find that
			if b.RemoteName != "" {
				for _, b2 := range c.Branches {
					if b.RemoteName == b2.Name {
						// Found the remote branch, prefer that
						return b2
					}
				}
			}
			// branch b had no remote branch, use local
			return b
		}
	}
	// Try find a branch with the display name
	for _, b := range c.Branches {
		if name == b.DisplayName {
			return b
		}
	}
	return nil
}

func (h *branchesService) isMergedDeletedBranch(repo *Repo, c *Commit) *Branch {
	if len(c.Branches) == 0 && len(c.Children) == 0 {
		// Commit has no branch, must be a deleted branch tip merged into some branch or unusual branch
		// Trying to ues parsed branch name from one of the merge children subjects e.g. Merge branch 'a' into develop
		name := h.branchNames.branchName(c.Id)
		if name != "" {
			// Managed to parse a branch name
			return repo.addNamedBranch(c, name)
		}

		// could not parse a name from any of the merge children, use id named branch
		return repo.addIdNamedBranch(c)
	}
	return nil
}

func (h *branchesService) isLocalRemoteBranch(c *Commit) *Branch {
	if len(c.Branches) == 2 {
		if c.Branches[0].IsRemote && c.Branches[0].Name == c.Branches[1].RemoteName {
			// remote and local branch, prefer remote
			return c.Branches[0]
		}
		if !c.Branches[0].IsRemote && c.Branches[0].RemoteName == c.Branches[1].Name {
			// local and remote branch, prefer remote
			return c.Branches[1]
		}
	}
	return nil
}
