package gitrepo

import (
	"fmt"
	"github.com/michael-reichenauer/gmc/utils/git"
)

type Tag struct {
	CommitID string
	TagName  string
}
type Repo struct {
	Commits    []*Commit
	commitById map[string]*Commit
	Branches   []*Branch
	Status     Status
	Tags       []Tag
	RepoPath   string
}

func newRepo() *Repo {
	return &Repo{commitById: make(map[string]*Commit)}
}

func (r *Repo) CommitByID(id string) *Commit {
	c, ok := r.commitById[id]
	if !ok {
		return partialLogCommit
	}
	return c

}

func (r *Repo) BranchByName(name string) (*Branch, bool) {
	for _, br := range r.Branches {
		if br.Name == name {
			return br, true
		}
	}
	return nil, false
}

func (r *Repo) Parent(commit *Commit, index int) (*Commit, bool) {
	if index >= len(commit.ParentIDs) {
		return nil, false
	}
	return r.CommitByID(commit.ParentIDs[index]), true
}

func (r *Repo) CurrentBranch() (*Branch, bool) {
	for _, br := range r.Branches {
		if br.IsCurrent {
			return br, true
		}
	}
	return nil, false
}

func (r *Repo) setGitCommits(gitCommits []git.Commit) {
	for _, gc := range gitCommits {
		commit := newCommit(gc)
		r.Commits = append(r.Commits, commit)
		r.commitById[commit.Id] = commit
	}

	// Set current commit if there is a current branch
	currentBranch, ok := r.CurrentBranch()
	if ok {
		currentCommit := r.CommitByID(currentBranch.TipID)
		currentCommit.IsCurrent = true
	}
}

func (r *Repo) setGitBranches(gitBranches []git.Branch) {
	for _, gb := range gitBranches {
		r.Branches = append(r.Branches, newBranch(gb))
	}
	// Set local name of all remote branches, that have a local branch as well
	for _, b := range r.Branches {
		if b.RemoteName != "" {
			// A local branch, try locate corresponding remote branch and set its local name property
			for _, rb := range r.Branches {
				if rb.Name == b.RemoteName {
					rb.LocalName = b.Name
					break
				}
			}
		}
	}
}

func (r *Repo) addMultiBranch(c *Commit) *Branch {
	b := &Branch{
		Name:          fmt.Sprintf("multi:%s", c.Sid),
		DisplayName:   fmt.Sprintf("multiple@%s", c.Sid),
		TipID:         c.Id,
		IsCurrent:     false,
		IsGitBranch:   false,
		IsMultiBranch: true,
		IsNamedBranch: false,
	}
	r.Branches = append(r.Branches, b)
	return b
}

func (r *Repo) addNamedBranch(c *Commit, branchName string) *Branch {
	b := &Branch{
		Name:          fmt.Sprintf("%s:%s", branchName, c.Sid),
		DisplayName:   branchName,
		TipID:         c.Id,
		IsCurrent:     false,
		IsGitBranch:   false,
		IsMultiBranch: false,
		IsNamedBranch: true,
	}
	r.Branches = append(r.Branches, b)
	return b
}
func (r *Repo) addIdNamedBranch(c *Commit) *Branch {
	b := &Branch{
		Name:          fmt.Sprintf("branch:%s", c.Sid),
		DisplayName:   fmt.Sprintf("branch@%s", c.Sid),
		TipID:         c.Id,
		IsCurrent:     false,
		IsGitBranch:   false,
		IsMultiBranch: false,
		IsNamedBranch: true,
	}
	r.Branches = append(r.Branches, b)
	return b
}
