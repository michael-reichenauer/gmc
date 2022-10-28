package augmented

import (
	"github.com/michael-reichenauer/gmc/utils"
	"github.com/michael-reichenauer/gmc/utils/linq"
	"github.com/samber/lo"
)

// - A commit has 0, 1 or 2 parent commits (down/below in list)
//   - 0 parents: The root (last) commit in the list is the root commit and has no parent.
//   - 1 parent:  A normal commit
//   - 2 parents: A merge commit
// - A commit can have 0 or multiple children (upp/above in list),
//   - 0 children are the tips commits of live git branches
//   - 1 child are normal commits, where the child commit only has one parent commit
//   - 2 or more children are commits from which one or more branches starts or ends (if merge child)
//   - 1 or more merge children are commit where those children are merge commits with a merge parent
// - Live branches are branches that have a git branch
// - Deleted branches have been merged into some commit, i.e for a commit with two parents
//   the first parent continuos the branch, while the second parent can be tip of a deleted branch,
//   or just a commit in a another branch
// - Merge commits often has a subject that indicate the name of the 2 branches;
//   Target: The current branch, Source the branch merged into the target.
// - Since the merge commit often has info of target and source branch names, the 2 parents commits
//   branches can also be reasonable determined.
// - Multiple branch tips can point to the same commit, that would make those
// - Main or master (legacy) branch are treated specially since these are considered backbone of the graph
// - Pull merge: Normally in git when a user does a pull merge, the remote branch is merged into
//   the local branch, where the remote branch tip will be the second parent of the local commit.
//   I think this is a bit strange and would confuse the branch graph. So if these commit messages
//   are detected, the order of the parents are switch so it looks like the local branch was
//   merged into the remote branch and not the reverse in normal git logs.

// Algorithm:
// 1 If commit only has one git branch, use that
// 2 if commit only has one local and corresponding remote branch, use the remote
// 3

// Default branch priority determines parent child branch relations.
var DefaultBranchPriority = []string{"origin/main", "main", "origin/master", "master"}

type branchesService struct {
	branchNames *branchNameParser
}

func newBranchesService() *branchesService {
	return &branchesService{branchNames: newBranchNameParser()}
}

func (h *branchesService) setBranchForAllCommits(repo *Repo) {
	branchesChildren := repo.MetaData.BranchesChildren

	h.setGitBranchTips(repo)
	h.setCommitBranchesAndChildren(repo)
	h.determineCommitBranches(repo, branchesChildren)
	h.mergeAmbiguousBranches(repo)
	h.determineBranchHierarchy(repo, branchesChildren)
}

// setGitBranchTips iterates all branches and
//   - add each branch to the branch tip commit branches,
//     Thus all branch tip commit knows the list of branches it belongs to,
//     Thus those tip commits parents will also know those branches
func (h *branchesService) setGitBranchTips(repo *Repo) {
	var invalidBranches []string
	for _, b := range repo.Branches {
		tip, ok := repo.TryGetCommitByID(b.TipID)
		if !ok {
			// A branch tip id, which commit id does not exist in the repo
			// Store that branch name so it can be removed from the list below
			invalidBranches = append(invalidBranches, b.Name)
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
	if len(invalidBranches) > 0 {
		repo.Branches = linq.Filter(repo.Branches, func(v *Branch) bool {
			return !linq.Contains(invalidBranches, v.Name)
		})
	}
}

// setCommitBranchesAndChildren iterates all commits and
//   - Swap commit parents order if the commit is a pull merge
//     In git, the first parent of pull merge commit is the local branch which I think is strange
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
	branchesChildren map[string][]string,
) {
	for _, c := range repo.Commits {
		branch := h.determineCommitBranch(repo, c, branchesChildren)
		c.Branch = branch
		c.addBranch(c.Branch)

		h.setMasterBackbone(c)
		c.Branch.BottomID = c.Id
	}
}

// determineCommitBranch determines the branch of a commit by analyzing the most likely
// candidate branch for a commit.
func (h *branchesService) determineCommitBranch(
	repo *Repo, c *Commit, branchesChildren map[string][]string,
) *Branch {
	if branch := h.hasOnlyOneBranch(c); branch != nil {
		// Commit only has one branch, it must have been an actual branch tip originally, use that
		return branch
	} else if branch := h.isLocalRemoteBranch(c); branch != nil {
		// Commit has only local and its remote branch, prefer remote remote branch
		return branch
	} else if branch := h.hasParentChildSetBranch(c, branchesChildren); branch != nil {
		// The commit has several possible branches, and one is set as parent of the others
		return branch
	} else if branch := h.hasChildrenPriorityBranch(c, branchesChildren); branch != nil {
		// The commit has several possible branches, but children
		return branch
	} else if branch := h.isSameChildrenBranches(c); branch != nil {
		// Commit has no branch but has 2 children with same branch
		return branch
	} else if branch := h.isMergedDeletedRemoteBranchTip(repo, c); branch != nil {
		// Commit has no branch and no children, but has a merge child, lets check if pull merger
		return branch
	} else if branch := h.isMergedDeletedBranchTip(repo, c); branch != nil {
		// Commit is a tip of a deleted branch, which was merged into a parent branch
		return branch
	} else if branch := h.hasOneChild(c); branch != nil {
		// Commit is middle commit in branch, use the only one child commit branch
		return branch
	} else if branch := h.hasOneChildWithLikelyBranch(c); branch != nil {
		// Commit has one child, which has a likely known branch, use same branch
		return branch
	} else if branch := h.hasPriorityBranch(c); branch != nil {
		// Commit, has several possible branches, and one is in the priority list, e.g. main, develop, ...
		return branch
	} else if branch := h.hasBranchNameInSubject(repo, c); branch != nil {
		// A branch name could be parsed form the commit subject or a child subject.
		return branch
	} else if branch := h.hasOnlyOneChild(c); branch != nil {
		// Commit has one child commit, use that child commit branch
		return branch
	} else if branch := h.isChildAmbiguousBranch(c); branch != nil {
		// one of the commit children is a ambiguous branch, reuse same ambiguous branch
		return branch
	}

	// Commit, has several possible branches, create a new ambiguous branch
	return repo.addAmbiguousBranch(c)
}

func (h *branchesService) hasOnlyOneBranch(c *Commit) *Branch {
	if len(c.Branches) == 1 {
		// Commit only has one branch, it must have been an actual branch tip originally, use that
		return c.Branches[0]
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

func (h *branchesService) hasParentChildSetBranch(commit *Commit, branchesChildren map[string][]string) *Branch {
	for _, b := range commit.Branches {
		childBranches := branchesChildren[b.BaseName()]
		if len(childBranches) == 0 {
			// This branch has no children branches
			continue
		}

		// assume c.Branch is parent of all other children branches (cc.Branch)
		assumeIsParent := true
		for _, bb := range commit.Branches {
			if b == bb || b.Name == bb.RemoteName {
				continue
			}
			if !utils.StringsContains(childBranches, bb.BaseName()) {
				// bb is not a child of b
				assumeIsParent = false
				break
			}
		}

		if assumeIsParent {
			// b was parent of all other branches
			return b
		}
	}

	return nil
}

// hasChildrenPriorityBranch iterates all children of commit and for each child
//   - If child has a branch which is a parent of all other children branches,
//     that branch is returned
func (h *branchesService) hasChildrenPriorityBranch(commit *Commit, branchesChildren map[string][]string) *Branch {
	if len(commit.Children) < 2 {
		return nil
	}

	for _, c := range commit.Children {
		childBranches := branchesChildren[c.Branch.BaseName()]
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
			if !utils.StringsContains(childBranches, cc.Branch.BaseName()) {
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

func (h *branchesService) isSameChildrenBranches(c *Commit) *Branch {
	if len(c.Branches) == 0 && len(c.Children) == 2 &&
		c.Children[0].Branch == c.Children[1].Branch {
		// Commit has no branch but has 2 children with same branch use that
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

func (h *branchesService) hasOneChild(c *Commit) *Branch {
	if len(c.Branches) == 0 && len(c.Children) == 1 {
		// Commit has no branch, but it has one child commit, use that child commit branch
		return c.Children[0].Branch
	}
	return nil
}

func (h *branchesService) hasOneChildWithLikelyBranch(c *Commit) *Branch {
	if len(c.Children) == 1 && c.Children[0].isLikely {
		// Commit has one child, which has a likely known branch, use same branch
		return c.Children[0].Branch
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

func (h *branchesService) hasBranchNameInSubject(repo *Repo, c *Commit) *Branch {
	if name := h.branchNames.branchName(c.Id); name != "" {
		// A branch name could be parsed form the commit subject or a child subject.
		// Lets use that as a branch name and also let children (commits above)
		// use that branch if they are an ambiguous branch
		var current *Commit
		branch := h.tryGetBranchFromName(c, name)
		if branch != nil && branch.BottomID != "" {
			// Found an existing branch with that name, set lowest known commit to the bottom
			// of that known branch
			current = repo.CommitByID(branch.BottomID)
		}

		if current == nil {
			// branch has no known last (bottom) commit, lets iterate upp (first child) as long
			// as commits are on an ambiguous branch
			for current = c; len(current.Children) == 1 && current.Children[0].Branch.IsAmbiguousBranch; current = current.Children[0] {
			}
		}

		if branch != nil {
			for ; current != nil && current != c.FirstParent; current = current.FirstParent {
				current.Branch = branch
				current.addBranch(branch)
				current.isLikely = true
			}

			return branch
		}
	}

	return nil
}

func (h *branchesService) hasOnlyOneChild(c *Commit) *Branch {
	// This does not work, since a commit could have multiple branch tips
	if len(c.Children) == 1 && len(c.MergeChildren) == 0 {
		// Commit has only one child, ensure commit has same branches
		child := c.Children[0]
		if len(c.Branches) != len(child.Branches) {
			// Number of branches have changed
			return nil
		}
		for i := 0; i < len(c.Branches); i++ {
			if c.Branches[i].Name != child.Branches[i].Name {
				return nil
			}
		}

		// Commit has one child commit, use that child commit branch
		return child.Branch
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

// setMasterBackbone, if the commit branch is one of the prioritized branches,
// that branch is added to the parent commit branches as well (inherited)
func (h *branchesService) setMasterBackbone(c *Commit) {
	if c.FirstParent == nil {
		// Reached the end of the repository
		return
	}

	if lo.Contains(DefaultBranchPriority, c.Branch.Name) {
		// main and develop are special and will make a "backbone" for other branches to depend on
		c.FirstParent.addBranch(c.Branch)
	}
}

// mergeAmbiguousBranches will merge commits of ambiguous branches into one of the child branches
func (h *branchesService) mergeAmbiguousBranches(repo *Repo) {
	ambiguousBranches := lo.Filter(repo.Branches, func(v *Branch, _ int) bool { return v.IsAmbiguousBranch })
	for _, b := range ambiguousBranches {
		tip := repo.CommitByID(b.TipID)

		// Determine the parent commit this branch was created from
		otherId := b.BottomID
		parentBranchCommit := repo.CommitByID(b.BottomID).FirstParent
		if parentBranchCommit != nil {
			otherId = parentBranchCommit.Id
		}

		// Find the tip of the ambiguous commits (and the next commit)
		var ambiguousTip *Commit
		var ambiguousSecond *Commit
		for c := tip; c != nil && c.Id != otherId; c = c.FirstParent {
			if c.Branch != b {
				// Still a normal branch commit (no longer part of the ambiguous branch)
				continue
			}

			// tip of the ambiguous commits
			ambiguousTip = c
			ambiguousSecond = c.FirstParent
			c.IsAmbiguousTip = true
			c.IsAmbiguous = true

			// Determine the most likely branch (branch of the oldest child)
			oldestChild := c.Children[0]
			childBranches := []*Branch{}
			for _, c := range c.Children {
				if c.AuthorTime.After(oldestChild.AuthorTime) {
					oldestChild = c
				}
				childBranches = append(childBranches, c.Branch)
			}
			c.Branch = oldestChild.Branch
			c.Branch.AmbiguousTipId = c.Id
			c.Branch.AmbiguousBranches = childBranches
			c.Branch.BottomID = c.Id
			break
		}

		// Set the branch of the rest of the ambiguous commits to same as the tip
		for c := ambiguousSecond; c != nil && c.Id != otherId; c = c.FirstParent {
			c.Branch = ambiguousTip.Branch
			c.Branch.BottomID = c.Id
			c.IsAmbiguous = true
		}

		// Removing the ambiguous branch (no longer needed)
		repo.Branches = lo.Filter(repo.Branches, func(v *Branch, _ int) bool { return v != b })
	}
}

func (h *branchesService) determineBranchHierarchy(repo *Repo, branchesChildren map[string][]string) {
	for _, b := range repo.Branches {
		bs := branchesChildren[b.BaseName()]
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
