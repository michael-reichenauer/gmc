package gitrepo

import (
	"github.com/michael-reichenauer/gmc/utils"
)

// Default branch priority determines parent child branch relations.
var DefaultBranchPriority = []string{"origin/main", "main", "origin/master", "master", "origin/develop", "develop", "origin/dev", "dev"}

type branchesService struct {
	branchNames *branchNameParser
}

func newBranchesService() *branchesService {
	return &branchesService{branchNames: newBranchNameParser()}
}

func (h *branchesService) setBranchForAllCommits(
	repo *Repo,
	branchesChildren map[string][]string) {
	h.setGitBranchTips(repo)
	h.setCommitBranchesAndChildren(repo)
	h.determineCommitBranches(repo, branchesChildren)
	h.determineBranchHierarchy(repo, branchesChildren)
}

// setGitBranchTips iterates all branches and
//   - add each branch to the branch tip commit branches,
//     Thus all branch tip commit knows the list of branches it belongs to,
//     Thus those tip commits parents will also know those branches
func (h *branchesService) setGitBranchTips(repo *Repo) {
	var missingBranches []string
	for _, b := range repo.Branches {
		tip, ok := repo.TryGetCommitByID(b.TipID)
		if !ok {
			// A branch tip id, which commit id does not exist in the repo
			// Store that branch name so it can be recmoved from the list below
			missingBranches = append(missingBranches, b.Name)
			continue
		}

		// Adding the branch to the branch tip commit
		tip.addBranch(b)
		tip.BranchTipNames = append(tip.BranchTipNames, b.Name)

		if b.IsCurrent {
			// Mark the current branch tip commit as the current commit
			tip.IsCurrent = true
		}
	}

	// If some branches do not have existing tip commit id,
	// Remove them from the list of repo.Branches
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

// setCommitBranchesAndChildren iterates all commits and
//   - Swap commit parents order if the commit is a pull merge
//     In git, the first parent of pull merge commit is the local branch which is a bit strange
//   - For the first parent commit, the child commit and its branches are set (inherited down)
//   - For the second parent (mergeParent), the commit is set but not its branches
func (h *branchesService) setCommitBranchesAndChildren(repo *Repo) {
	for _, c := range repo.Commits {
		h.branchNames.parseCommit(c)
		if len(c.ParentIDs) == 2 && h.branchNames.isPullMerge(c) {
			// if the commit is a pull merger, we do switch the order of parents
			// So the first parent is the remote branch and second parent the local branch
			c.ParentIDs = []string{c.ParentIDs[1], c.ParentIDs[0]}
		}

		firstParent, ok := repo.Parent(c, 0)
		if ok {
			c.FirstParent = firstParent
			firstParent.Children = append(firstParent.Children, c)
			firstParent.ChildIDs = append(firstParent.ChildIDs, c.Id)
			// Adding the child branches to the parent branches (inherited down)
			firstParent.addBranches(c.Branches)
		}

		mergeParent, ok := repo.Parent(c, 1)
		if ok {
			c.MergeParent = mergeParent
			mergeParent.MergeChildren = append(mergeParent.MergeChildren, c)
			mergeParent.ChildIDs = append(mergeParent.ChildIDs, c.Id)
			// Note: merge parent do not inherit child branches
		}
	}
}

// determineCommitBranches iterates all commits and for each commit
// - Determine the branch for the commit
// - If the commit has a prioritized branch, the first parent inherits that as well
// - Adjust the commit branch bottom id to know/forward last known commit of a branch
func (h *branchesService) determineCommitBranches(
	repo *Repo,
	branchesChildren map[string][]string) {
	for _, c := range repo.Commits {
		h.determineCommitBranch(repo, c, branchesChildren)
		h.setMasterBackbone(c)
		c.Branch.BottomID = c.Id
	}
}

// determineCommitBranch determines the branch of a commit by analyzing the most likely candidate branch
func (h *branchesService) determineCommitBranch(
	repo *Repo, c *Commit, branchesChildren map[string][]string) {
	if len(c.Branches) == 1 {
		// Commit only has one branch, it must have been an actual branch tip originally, use that
		c.Branch = c.Branches[0]
		return
	}

	if branch := h.isLocalRemoteBranch(c); branch != nil {
		// Commit has only local and its remote branch, prefer remote remote branch
		c.Branch = branch
		return
	}

	if branch := h.hasChildrenPriorityBranch(c, branchesChildren); branch != nil {
		// The commit has several possible branches, but children
		c.Branch = branch
		c.addBranch(c.Branch)
		return
	}

	if branch := h.isSameChildrenBranches(repo, c); branch != nil {
		c.Branch = branch
		c.addBranch(c.Branch)
		return
	}

	if branch := h.isMergedDeletedRemoteBranchTip(repo, c); branch != nil {
		c.Branch = branch
		c.addBranch(c.Branch)
		return
	}

	if branch := h.isMergedDeletedBranchTip(repo, c); branch != nil {
		// Commit is a tip of a deleted branch, which was merged into a parent branch
		c.Branch = branch
		c.addBranch(c.Branch)
		return
	}

	if branch := h.hasOneChild(repo, c); branch != nil {
		// Commit is middle commit in branch, use the only one child commit branch
		c.Branch = branch
		c.addBranch(c.Branch)
		return
	}

	if len(c.Children) == 1 && c.Children[0].isLikely {
		// Commit has one child, which has a likely known branch, use same branch
		c.Branch = c.Children[0].Branch
		c.isLikely = true
		return
	}

	if branch := h.hasPriorityBranch(c); branch != nil {
		// Commit, has several possible branches, and one is in the priority list, e.g. main, develop, ...
		c.Branch = branch
		return
	}

	if name := h.branchNames.branchName(c.Id); name != "" {
		// A branch name could be parsed form the commit subject or a child subject.
		// Lets use that as a branch name and also let children use that branch if they only are ambiguous branch
		branch := h.tryGetBranchFromName(c, name)
		var current *Commit
		if branch != nil && branch.BottomID != "" {
			current = repo.CommitByID(branch.BottomID)
		}
		if current == nil {
			for current = c; len(current.Children) == 1 && current.Children[0].Branch.IsAmbiguousBranch; current = current.Children[0] {
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

	if branch := h.isChildAmbiguousBranch(c); branch != nil {
		// one of the commit children is a ambiguous branch, reuse same ambiguous branch
		c.Branch = branch
		c.addBranch(c.Branch)
		return
	}

	// Commit, has several possible branches, create a new ambiguous branch
	c.Branch = repo.addAmbiguousBranch(c)
	c.addBranch(c.Branch)
}

// hasChildrenPriorityBranch iterates all children of commit and for each child
//   - If child has a branch which is a parent of all other children branches,
//     that branch is returned
func (h *branchesService) hasChildrenPriorityBranch(commit *Commit, branchesChildren map[string][]string) *Branch {
	if len(commit.Children) < 2 {
		return nil
	}

	for _, c := range commit.Children {
		childBranches := branchesChildren[c.Branch.Name]
		if len(childBranches) == 0 {
			// This child branch has no children branches
			continue
		}

		// assume c.Branch is parent of all other children branches (cc.Branch)
		assumeIsParent := true
		for _, cc := range commit.Children {
			if c == cc {
				continue
			}
			if !utils.StringsContains(childBranches, cc.Branch.Name) {
				// cc.Branch is not a child of c.Branch
				assumeIsParent = false
				break
			}
		}

		if assumeIsParent {
			// c.Branch was parent of all other children branches
			return c.Branch
		}
	}

	return nil
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

func (h *branchesService) isChildAmbiguousBranch(c *Commit) *Branch {
	for _, cc := range c.Children {
		if cc.Branch != nil && cc.Branch.IsAmbiguousBranch {
			// one of the commit children is a ambiguous branch
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

func (h *branchesService) isSameChildrenBranches(repo *Repo, c *Commit) *Branch {
	if len(c.Branches) == 0 && len(c.Children) == 2 &&
		c.Children[0].Branch == c.Children[1].Branch {
		// Commit has no branch and no children, but has 2 children with same branch use that
		return c.Children[0].Branch
	}
	return nil
}

func (h *branchesService) isMergedDeletedRemoteBranchTip(repo *Repo, c *Commit) *Branch {
	if len(c.Branches) == 0 && len(c.Children) == 0 && len(c.MergeChildren) == 1 {
		// Commit has no branch and no children, but has a merge child, lets check if pull merger
		// Trying to use parsed branch name from one of the merge children subjects e.g. Merge branch 'a' into develop
		name := h.branchNames.branchName(c.Id)
		if name != "" {
			// Managed to parse a branch name
			mergeChildBranch := c.MergeChildren[0].Branch
			if name == mergeChildBranch.DisplayName {
				return mergeChildBranch
			}

			return repo.addNamedBranch(c, name)
		}

		// could not parse a name from any of the merge children, use id named branch
		return repo.addIdNamedBranch(c)
	}
	return nil
}

func (h *branchesService) isMergedDeletedBranchTip(repo *Repo, c *Commit) *Branch {
	if len(c.Branches) == 0 && len(c.Children) == 0 {
		// Commit has no branch, must be a deleted branch tip merged into some branch or unusual branch
		// Trying to use parsed branch name from one of the merge children subjects e.g. Merge branch 'a' into develop
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

func (h *branchesService) hasOneChild(repo *Repo, c *Commit) *Branch {
	if len(c.Branches) == 0 && len(c.Children) == 1 {
		// Commit has no branch, but it has one child commit, use that child commit branch
		return c.Children[0].Branch
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

// setMasterBackbone, if the commit branch is one of the prioritized branches,
// that branch is added to the parent commit branches as well (inherited)
func (h *branchesService) setMasterBackbone(c *Commit) {
	if c.FirstParent == nil {
		// Reached the end of the repository
		return
	}

	if utils.StringsContains(DefaultBranchPriority, c.Branch.Name) {
		// main and develop are special and will make a "backbone" for other branches to depend on
		c.FirstParent.addBranch(c.Branch)
	}
}

func (h *branchesService) determineBranchHierarchy(repo *Repo, branchesChildren map[string][]string) {
	for _, b := range repo.Branches {
		bs, _ := branchesChildren[b.Name]
		b.IsSetAsParent = len(bs) > 0

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
