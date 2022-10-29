package augmented

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/michael-reichenauer/gmc/utils/git"
	"github.com/michael-reichenauer/gmc/utils/linq"
	"github.com/michael-reichenauer/gmc/utils/log"
)

type MetaData struct {
	BranchesChildren map[string][]string
}

type Repo struct {
	Commits    []*Commit
	commitById map[string]*Commit
	Branches   []*Branch
	Status     Status
	Tags       []Tag
	RepoPath   string
	MetaData   MetaData
}

// augmented
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
	// For repositories with a lot of commits, only the latest 'partialMax' number of commits
	// are used, i.w. partial commits, which should have parents, but they are unknown
	isPartialPossible := len(gitCommits) >= partialMax
	isPartialNeeded := false
	commits := make([]*Commit, len(gitCommits), len(gitCommits)+10)

	// Iterate git commits in reverse
	for i := len(gitCommits) - 1; i >= 0; i-- {
		gc := gitCommits[i]
		commit := newGitCommit(gc)

		if isPartialPossible {
			// The repo was truncated, check if commits have missing parents, which will be set
			// to a virtual/fake "partial commit"
			if len(commit.ParentIDs) == 1 {
				// Not a merge commit but check if parent is missing and need a partial commit parent
				if _, ok := r.commitById[commit.ParentIDs[0]]; !ok {
					isPartialNeeded = true
					commit.ParentIDs = []string{git.PartialLogCommitID}
				}
			}

			if len(commit.ParentIDs) == 2 {
				// Merge commit, check if parents are missing and need a partial commit parent
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
		// Add a virtual/fake partial commit, which some commits will have as a parent
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
	r.Branches = linq.Map(gitBranches, func(b git.Branch) *Branch { return newGitBranch(b) })

	// Set local name of all remote branches, that have a corresponding local branch as well
	// Unset RemoteName of local branch if n corresponding remote branch (deleted on remote server)
	for _, b := range r.Branches {
		if b.RemoteName != "" {
			remoteBranch, ok := linq.Find(r.Branches, func(v *Branch) bool { return v.Name == b.RemoteName })
			if ok {
				// Corresponding remote branch, set local branch name property
				remoteBranch.LocalName = b.Name
			} else {
				// No remote corresponding remote branch, unset property
				b.RemoteName = ""
			}
		}
	}
}

func (r *Repo) addAmbiguousBranch(c *Commit) *Branch {
	b := newAmbiguousBranch(c.Id)
	for _, cc := range c.Children {
		b.AmbiguousBranches = append(b.AmbiguousBranches, cc.Branch)
	}
	r.Branches = append(r.Branches, b)
	return b
}

func (r *Repo) addNamedBranch(c *Commit, branchName string) *Branch {
	b := newNamedBranch(c.Id, branchName)
	r.Branches = append(r.Branches, b)
	return b
}

func (r *Repo) addIdNamedBranch(c *Commit) *Branch {
	b := newUnnamedBranch(c.Id)
	r.Branches = append(r.Branches, b)
	return b
}

func toMetaData(metaDataText string) MetaData {
	if metaDataText == "" {
		return defaultMetaData()
	}

	var metaData MetaData
	err := json.Unmarshal([]byte(metaDataText), &metaData)
	if err != nil {
		log.Warnf("Failed to parse %q, %v", metaDataText, err)
		return defaultMetaData()
	}

	if metaData.BranchesChildren == nil {
		metaData.BranchesChildren = make(map[string][]string)
	}
	return metaData
}

func defaultMetaData() MetaData {
	return MetaData{BranchesChildren: make(map[string][]string)}
}
