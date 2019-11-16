package gitmodel

import (
	"github.com/michael-reichenauer/gmc/utils/git"
	"github.com/michael-reichenauer/gmc/utils/log"
	"strings"
)

var defaultBranchPrio = []string{"master:local", "develop:local"}

type Handler struct {
	gitRepo *git.Repo
	repo    *Repo
	err     error
}

func NewModel(repoPath string) *Handler {
	return &Handler{
		gitRepo: git.NewRepo(repoPath),
		repo:    newRepo(),
	}
}

func (h *Handler) GetRepo() Repo {
	return *h.repo
}

func (h *Handler) Load() {
	err := h.initRepo()
	if err != nil {
		h.err = err
	}
}
func (h *Handler) initRepo() error {
	h.repo = newRepo()
	gitCommits, err := h.gitRepo.GetLog()
	if err != nil {
		return err
	}
	gitBranches, err := h.gitRepo.GetBranches()
	if err != nil {
		return err
	}

	h.repo.setGitCommits(gitCommits)
	h.repo.setGitBranches(gitBranches)

	h.setGitBranchTips()
	h.setCommitBranchesAndChildren()

	h.determineCommitBranches()
	return nil
}

func (h *Handler) setCommitBranchesAndChildren() {
	for _, c := range h.repo.Commits {
		parent, ok := h.repo.Parent(c, 0)
		if ok {
			c.Parent = parent
			c.Parent.Children = append(c.Parent.Children, c)
			parent.addBranches(c.Branches)
			parent.ChildIDs = append(parent.ChildIDs, c.Id)
		}

		mergeParent, ok := h.repo.Parent(c, 1)
		if ok {
			c.MergeParent = mergeParent
			c.MergeParent.MergeChildren = append(c.MergeParent.MergeChildren, c)
			parent.ChildIDs = append(parent.ChildIDs, c.Id)
		}
	}
}

func (h *Handler) setGitBranchTips() {
	for _, b := range h.repo.Branches {
		h.repo.CommitById(b.TipID).addBranch(b)
		if b.IsCurrent {
			h.repo.CommitById(b.TipID).IsCurrent = true
		}
	}
}

func (h *Handler) determineCommitBranches() {
	for _, c := range h.repo.Commits {
		h.determineBranch(c)
	}
}

func (h *Handler) determineBranch(c *Commit) {
	if strings.HasPrefix(c.Id, "81e1a9f") {
		log.Infof("")
	}
	if c.Branch != nil {
		// Commit already knows its branch, e.g. deleted merged branch
		return
	}

	if len(c.Branches) == 1 {
		// Commit only has one branch, use that
		c.Branch = c.Branches[0]
		return
	}

	if len(c.Branches) == 0 && len(c.Children) == 0 {
		// Commit has no branch, must be a deleted branch tip merged into some branch or unusual branch
		// Trying to parse a branch name from one of the merge children subjects e.g. Merge branch 'a' into develop
		for _, mc := range c.MergeChildren {
			from, _ := h.parseMergeBranchNames(mc.Subject)
			if from != "" {
				// Managed to parse a branch name
				c.Branch = h.repo.AddNamedBranch(c, from)
				c.Branches = append(c.Branches, c.Branch)
				return
			}
		}
		// could not parse a name from any of the merge children, use id named branch
		c.Branch = h.repo.AddIdNamedBranch(c)
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
	for _, bp := range defaultBranchPrio {
		for _, cb := range c.Branches {
			if bp == cb.ID {
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
	c.Branch = h.repo.AddMultiBranch(c)
	c.Branches = append(c.Branches, c.Branch)
}

func (h *Handler) parseMergeBranchNames(subject string) (from string, into string) {
	if strings.HasPrefix(subject, "Merge branch '") {
		ei := strings.LastIndex(subject, "'")
		if ei > 14 {
			from = subject[14:ei]
			if strings.HasPrefix(subject[ei:], "' into ") {
				into = subject[ei+7:]
			}
		}
	}
	return
}
