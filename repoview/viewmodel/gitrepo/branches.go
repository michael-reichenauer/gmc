package gitrepo

// Default branch priority determines parent child branch relations
var DefaultBranchPriority = []string{"origin/master", "master", "origin/develop", "develop"}

type branches struct {
	branchNames *branchNames
}

func newBranches() *branches {
	return &branches{branchNames: newBranchNames()}
}

func (h *branches) setBranchForAllCommits(repo *Repo) {
	h.setGitBranchTips(repo)
	h.setCommitBranchesAndChildren(repo)
	h.determineCommitBranches(repo)
	h.determineBranchHierarchy(repo)

}

func (h *branches) setGitBranchTips(repo *Repo) {
	for _, b := range repo.Branches {
		repo.CommitById[b.TipID].addBranch(b)
		if b.IsCurrent {
			repo.CommitById[b.TipID].IsCurrent = true
		}
	}
}

func (h *branches) setCommitBranchesAndChildren(repo *Repo) {
	for _, c := range repo.Commits {
		parent, ok := repo.Parent(c, 0)
		if ok {
			c.Parent = parent
			c.Parent.Children = append(c.Parent.Children, c)
			parent.addBranches(c.Branches)
			parent.ChildIDs = append(parent.ChildIDs, c.Id)
		}

		mergeParent, ok := repo.Parent(c, 1)
		if ok {
			c.MergeParent = mergeParent
			c.MergeParent.MergeChildren = append(c.MergeParent.MergeChildren, c)
			parent.ChildIDs = append(parent.ChildIDs, c.Id)
		}
	}
}

func (h *branches) determineBranchHierarchy(repo *Repo) {
	for _, b := range repo.Branches {
		if b.BottomID == "" {
			b.BottomID = b.TipID
		}

		bottom := repo.CommitById[b.BottomID]
		if bottom.Branch != b {
			// the tip does not own the tip commit, i.e. a branch pointer to another branch
			b.ParentBranch = bottom.Branch
		} else if bottom.Parent != nil {
			b.ParentBranch = bottom.Parent.Branch
		}
	}
}

func (h *branches) determineCommitBranches(repo *Repo) {
	for _, c := range repo.Commits {
		h.branchNames.parseCommit(c)

		h.determineBranch(repo, c)
		c.Branch.BottomID = c.Id
	}
}

func (h *branches) determineBranch(repo *Repo, c *Commit) {
	if c.Branch != nil {
		// Commit already knows its branch
		panic("Commit already knows its branch") // ##############?????????
		return
	}

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
		// commit has no known branches (middle commit in deleted branch), but has one child, use that branch
		c.Branch = c.Children[0].Branch
		c.addBranch(c.Branch)
		return
	}

	if len(c.Children) == 1 && c.Children[0].IsLikely {
		c.Branch = c.Children[0].Branch
		c.IsLikely = true
		return
	}

	if branch := h.hasPriorityBranch(c); branch != nil {
		// Commit, has many possible branches, check if one is in the priority list, e.g. master, develop, ...
		c.Branch = branch
		return
	}

	if name := h.branchNames.branchName(c.Id); name != "" {
		// The commit branch name could be parsed form the subject (or a child subject)
		// Lets use that as a branch and also let children use that branch if they only are multi branch
		var current *Commit
		for current = c; len(current.Children) == 1 && current.Children[0].Branch.IsMultiBranch; current = current.Children[0] {
		}
		branch := h.tryGetBranchFromName(c, name)
		if branch == nil {
			branch = repo.AddNamedBranch(current, name)
		}
		for ; current != c.Parent; current = current.Parent {
			current.Branch = branch
			current.IsLikely = true
			current.addBranch(branch)
		}
		c.Branch.BottomID = c.Id
		return
	}

	if branch := h.isChildMultiBranch(c); branch != nil {
		// one of the commit children is a multi branch, reuse
		c.Branch = branch
		c.addBranch(c.Branch)
		return
	}

	// Commit, has several possible branches, create a new multi branch
	c.Branch = repo.AddMultiBranch(c)
	c.addBranch(c.Branch)
}

func (h *branches) hasPriorityBranch(c *Commit) *Branch {
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
func (h *branches) isChildMultiBranch(c *Commit) *Branch {
	for _, cc := range c.Children {
		if cc.Branch.IsMultiBranch {
			// one of the commit children is a multi branch
			return cc.Branch
		}
	}
	return nil
}

func (h *branches) tryGetBranchFromName(c *Commit, name string) *Branch {
	for _, b := range c.Branches {
		if name == b.DisplayName {
			return b
		}
	}
	return nil
}

func (h *branches) isMergedDeletedBranch(repo *Repo, c *Commit) *Branch {
	if len(c.Branches) == 0 && len(c.Children) == 0 {
		// Commit has no branch, must be a deleted branch tip merged into some branch or unusual branch
		// Trying to ues parsed branch name from one of the merge children subjects e.g. Merge branch 'a' into develop
		name := h.branchNames.branchName(c.Id)
		if name != "" {
			// Managed to parse a branch name
			return repo.AddNamedBranch(c, name)
		}

		// could not parse a name from any of the merge children, use id named branch
		return repo.AddIdNamedBranch(c)
	}
	return nil
}

func (h *branches) isLocalRemoteBranch(c *Commit) *Branch {
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
