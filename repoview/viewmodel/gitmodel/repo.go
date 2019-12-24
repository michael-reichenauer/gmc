package gitmodel

import (
	"fmt"
	"github.com/michael-reichenauer/gmc/utils/git"
)

type Repo struct {
	Commits    []*Commit
	CommitById map[string]*Commit
	Branches   []*Branch
	Status     Status
	RepoPath   string
}

func newRepo() *Repo {
	r := &Repo{
		//Commits:              []*Commit{},
		CommitById: make(map[string]*Commit),
		//Branches:             []*Branch{},
	}
	return r
}

//func (r *Repo) CommitById(id string) *Commit {
//	return r.commitById[id]
//}
func (r *Repo) setGitCommits(gitCommits []git.Commit) {
	//r.Commits = []*Commit{}
	//r.CommitById = make(map[string]*Commit)

	for _, gc := range gitCommits {
		commit := newCommit(gc)
		r.Commits = append(r.Commits, commit)
		r.CommitById[commit.Id] = commit
	}

	// Set current commit if there is a current branch
	currentBranch, ok := r.CurrentBranch()
	if ok {
		currentCommit := r.CommitById[currentBranch.TipID]
		currentCommit.IsCurrent = true
	}
}

func (r *Repo) setGitBranches(gitBranches []git.Branch) {
	for _, gb := range gitBranches {
		r.Branches = append(r.Branches, newBranch(gb))
	}
}

func (r *Repo) BranchByName(name string) (*Branch, bool) {
	for _, br := range r.Branches {
		if br.Name == name {
			return br, true
		}
	}
	return nil, false
}
func (r *Repo) LocalBranchByRemoteName(remoteName string) (*Branch, bool) {
	for _, br := range r.Branches {
		if br.RemoteName == remoteName {
			return br, true
		}
	}
	return nil, false
}
func (r *Repo) BranchByID(id string) (*Branch, bool) {
	for _, br := range r.Branches {
		if br.Name == id {
			return br, true
		}
	}
	return nil, false
}

func (r *Repo) CurrentBranch() (*Branch, bool) {
	for _, br := range r.Branches {
		if br.IsCurrent {
			return br, true
		}
	}
	return nil, false
}

func (r *Repo) Parent(commit *Commit, index int) (*Commit, bool) {
	if index >= len(commit.ParentIDs) {
		return nil, false
	}
	c, ok := r.CommitById[commit.ParentIDs[index]]
	return c, ok
}

func (r *Repo) AddMultiBranch(c *Commit) *Branch {
	b := &Branch{
		Name:          fmt.Sprintf("multi:%s", c.Sid),
		DisplayName:   fmt.Sprintf("multi:%s", c.Sid),
		TipID:         c.Id,
		IsCurrent:     false,
		IsGitBranch:   false,
		IsMultiBranch: true,
		IsNamedBranch: false,
	}
	r.Branches = append(r.Branches, b)
	return b
}

func (r *Repo) AddNamedBranch(c *Commit, branchName string) *Branch {
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
func (r *Repo) AddIdNamedBranch(c *Commit) *Branch {
	b := &Branch{
		Name:          fmt.Sprintf("branch:%s", c.Sid),
		DisplayName:   fmt.Sprintf("branch:%s", c.Sid),
		TipID:         c.Id,
		IsCurrent:     false,
		IsGitBranch:   false,
		IsMultiBranch: false,
		IsNamedBranch: true,
	}
	r.Branches = append(r.Branches, b)
	return b
}
