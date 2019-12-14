package gitmodel

import (
	"github.com/michael-reichenauer/gmc/utils/git"
	"sync"
)

var DefaultBranchPrio = []string{"origin/master", "master", "origin/develop", "develop"}

type Handler struct {
	gitRepo     *git.Repo
	branchNames *branchNamesHandler
	lock        sync.Mutex
	currentRepo *Repo
	err         error
}

func NewModel(repoPath string) *Handler {
	return &Handler{
		gitRepo:     git.NewRepo(repoPath),
		branchNames: newBranchNamesHandler(),
		currentRepo: newRepo(),
	}
}

func (h *Handler) GetRepo() Repo {
	h.lock.Lock()
	defer h.lock.Unlock()
	return *h.currentRepo
}

func (h *Handler) Load() {
	err := h.refreshRepo()
	if err != nil {
		h.err = err
	}
}

func (h *Handler) refreshRepo() error {
	repo := newRepo()
	gitCommits, err := h.gitRepo.GetLog()
	if err != nil {
		return err
	}
	gitBranches, err := h.gitRepo.GetBranches()
	if err != nil {
		return err
	}
	gitStatus, err := h.gitRepo.GetStatus()
	if err != nil {
		return err
	}

	repo.Status = newStatus(gitStatus)

	repo.setGitCommits(gitCommits)
	repo.setGitBranches(gitBranches)

	h.setGitBranchTips(repo)
	h.setCommitBranchesAndChildren(repo)
	h.determineCommitBranches(repo)
	h.determineBranchHierarchy(repo)

	h.lock.Lock()
	h.currentRepo = repo
	h.lock.Unlock()
	return nil
}

func (h *Handler) setCommitBranchesAndChildren(repo *Repo) {
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

func (h *Handler) setGitBranchTips(repo *Repo) {
	for _, b := range repo.Branches {
		repo.CommitById(b.TipID).addBranch(b)
		if b.IsCurrent {
			repo.CommitById(b.TipID).IsCurrent = true
		}
	}
}

func (h *Handler) determineCommitBranches(repo *Repo) {
	for _, c := range repo.Commits {
		h.branchNames.parseCommit(c)

		h.determineBranch(repo, c)
		c.Branch.BottomID = c.Id // ##############?????????
	}
}

func (h *Handler) determineBranch(repo *Repo, c *Commit) {
	if c.Branch != nil {
		// Commit already knows its branch, e.g. deleted merged branch
		return
	}

	if len(c.Branches) == 1 {
		// Commit only has one branch, use that
		c.Branch = c.Branches[0]
		return
	}

	if len(c.Branches) == 2 {
		if c.Branches[0].IsRemote && c.Branches[0].Name == c.Branches[1].RemoteName {
			// remote and local branch, prefer remote
			c.Branch = c.Branches[0]
			return
		}
		if !c.Branches[0].IsRemote && c.Branches[0].RemoteName == c.Branches[1].Name {
			// local and remote branch, prefer remote
			c.Branch = c.Branches[1]
			return
		}
	}

	if len(c.Branches) == 0 && len(c.Children) == 0 {
		// Commit has no branch, must be a deleted branch tip merged into some branch or unusual branch
		// Trying to ues parsed branch name from one of the merge children subjects e.g. Merge branch 'a' into develop
		name := h.branchNames.branchName(c.Id)
		if name != "" {
			// Managed to parse a branch name
			c.Branch = repo.AddNamedBranch(c, name)
			c.Branches = append(c.Branches, c.Branch)
			return
		}

		// could not parse a name from any of the merge children, use id named branch
		c.Branch = repo.AddIdNamedBranch(c)
		c.Branches = append(c.Branches, c.Branch)
		return
	}
	if len(c.Branches) == 0 && len(c.Children) == 1 {
		// commit has no known branches, but has one child, use that branch
		c.Branch = c.Children[0].Branch
		c.Branches = append(c.Branches, c.Branch)
		return
	}

	// Commit, has many possible branches, check if one is in the priority list, e.g. master, develop, ...
	for _, bp := range DefaultBranchPrio {
		for _, cb := range c.Branches {
			if bp == cb.Name {
				c.Branch = cb
				return
			}
		}
	}

	for _, cc := range c.Children {
		if cc.Branch.IsMultiBranch {
			// one of the commit children is a multi branch, reuse
			c.Branch = cc.Branch
			c.Branches = append(c.Branches, c.Branch)
			return
		}
	}

	// Commit, has many possible branches and many children, create a new multi branch
	c.Branch = repo.AddMultiBranch(c)
	c.Branches = append(c.Branches, c.Branch)
}

func (h *Handler) determineBranchHierarchy(repo *Repo) {
	for _, b := range repo.Branches {
		if b.BottomID == "" {
			b.BottomID = b.TipID
		}

		bottom := repo.commitById[b.BottomID]
		if bottom.Branch != b {
			// the tip does not own the tip commit, i.e. a branch pointer to another branch
			b.ParentBranch = bottom.Branch
		} else if bottom.Parent != nil {
			b.ParentBranch = bottom.Parent.Branch
		}
	}
}
