package viewmodel

import (
	"time"
)

type ViewPort struct {
	repo               *repo
	Commits            []Commit
	TotalCommits       int
	FirstIndex         int
	LastIndex          int
	CurrentCommitIndex int
	GraphWidth         int
	SelectedBranch     Branch
	CurrentBranchName  string
	First              int
	Last               int
	Selected           int
	RepoPath           string
	UncommittedChanges int
}

type Commit struct {
	ID         string
	SID        string
	Subject    string
	Message    string
	Author     string
	AuthorTime time.Time
	ParentIDs  []string
	IsCurrent  bool
	Branch     Branch
	Index      int
	Graph      []GraphColumn
	IsMore     bool
}

type Branch struct {
	Name          string
	DisplayName   string
	Index         int
	IsMultiBranch bool
	RemoteName    string
	IsRemote      bool
}

func newViewPort(repo *repo, first, last, selected int) ViewPort {
	size := last - first
	if last >= len(repo.Commits) {
		last = len(repo.Commits) - 1
	}
	first = last - size
	if first < 0 {
		first = 0
	}

	if selected > last {
		selected = last
	}
	if selected < first {
		selected = first
	}
	var selectedBranch Branch
	if selected < len(repo.Commits) {
		selectedBranch = toBranch(repo.Commits[selected].Branch)
	}

	return ViewPort{
		Commits:            toCommits(repo, first, last),
		FirstIndex:         first,
		LastIndex:          last,
		TotalCommits:       len(repo.Commits),
		CurrentBranchName:  repo.CurrentBranchName,
		GraphWidth:         len(repo.Branches) * 2,
		SelectedBranch:     selectedBranch,
		repo:               repo,
		First:              first,
		Last:               last,
		Selected:           selected,
		RepoPath:           repo.gmRepo.RepoPath,
		UncommittedChanges: repo.gmStatus.AllChanges(),
	}
}

func toCommits(repo *repo, first int, last int) []Commit {
	if first == -1 {
		return []Commit{}
	}
	commits := make([]Commit, last-first+1)
	for i := first; i <= last; i++ {
		commits[i-first] = toCommit(repo.Commits[i])
	}
	return commits
}

func toCommit(c *commit) Commit {
	return Commit{
		ID:         c.ID,
		SID:        c.SID,
		Subject:    c.Subject,
		Message:    c.Message,
		Author:     c.Author,
		AuthorTime: c.AuthorTime,
		ParentIDs:  c.ParentIDs,
		IsCurrent:  c.IsCurrent,
		Branch:     toBranch(c.Branch),
		Graph:      c.graph,
		Index:      c.Index,
		IsMore:     c.IsMore,
	}
}

func toBranch(b *branch) Branch {
	return Branch{
		Name:          b.name,
		DisplayName:   b.displayName,
		Index:         b.index,
		IsMultiBranch: b.isMultiBranch,
		RemoteName:    b.remoteName,
		IsRemote:      b.isRemote,
	}
}
