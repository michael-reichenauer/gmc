package gitrepo

import (
	"fmt"
	"github.com/michael-reichenauer/gmc/utils/git"
	"github.com/michael-reichenauer/gmc/utils/log"
	"strings"
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

func (r Repo) SearchCommits(searchText string) []*Commit {
	searchText = strings.ToLower(searchText)

	// Remember tag commits which contain the search string
	tagCommits := make(map[string]bool)
	for _, tag := range r.Tags {
		if textContainsText(tag.TagName, searchText) {
			tagCommits[tag.CommitID] = true
		}
	}

	var commits []*Commit
	for _, c := range r.Commits {
		if c.containsText(searchText) {
			// The commit contained the search string
			commits = append(commits, c)
			continue
		}
		if _, ok := tagCommits[c.Id]; ok {
			// Tags for the commit contained the search string
			commits = append(commits, c)
			continue
		}
	}
	return commits
}

func (r *Repo) CommitByID(id string) *Commit {
	c, ok := r.commitById[id]
	if !ok {
		panic(log.Fatal(fmt.Errorf("failed to find commit %s", id)))
	}
	return c
}

func (r *Repo) TryGetCommitByID(id string) (*Commit, bool) {
	c, ok := r.commitById[id]
	return c, ok
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
	isPartialPossible := len(gitCommits) >= partialMax
	commits := make([]*Commit, len(gitCommits), len(gitCommits)+10)
	isPartialNeeded := false
	for i := len(gitCommits) - 1; i > -1; i-- {
		gc := gitCommits[i]
		commit := newGitCommit(gc)
		if isPartialPossible {
			if len(commit.ParentIDs) == 1 {
				if _, ok := r.commitById[commit.ParentIDs[0]]; !ok {
					isPartialNeeded = true
					commit.ParentIDs = []string{git.PartialLogCommitID}
				}
			}
			if len(commit.ParentIDs) == 2 {
				if _, ok := r.commitById[commit.ParentIDs[0]]; !ok {
					isPartialNeeded = true
					commit.ParentIDs = []string{git.PartialLogCommitID, commit.ParentIDs[1]}
				}
				if _, ok := r.commitById[commit.ParentIDs[1]]; !ok {
					isPartialNeeded = true
					commit.ParentIDs = []string{commit.ParentIDs[0], git.PartialLogCommitID}
				}
			}
		}

		commits[i] = commit
		r.commitById[commit.Id] = commit
	}
	r.Commits = commits

	if isPartialNeeded {
		pc := newPartialLogCommit()
		r.Commits = append(r.Commits, pc)
		r.commitById[pc.Id] = pc
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
		r.Branches = append(r.Branches, newGitBranch(gb))
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
	b := newMultiBranch(c.Id)
	r.Branches = append(r.Branches, b)
	return b
}

func (r *Repo) addNamedBranch(c *Commit, branchName string) *Branch {
	b := newNamedBranch(c.Id, branchName)
	r.Branches = append(r.Branches, b)
	return b
}

func (r *Repo) addIdNamedBranch(c *Commit) *Branch {
	b := newBranch(c.Id)
	r.Branches = append(r.Branches, b)
	return b
}
