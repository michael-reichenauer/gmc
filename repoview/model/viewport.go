package model

import (
	"time"
)

type ViewPort struct {
	Commits            []Commit
	TotalCommits       int
	FirstIndex         int
	LastIndex          int
	CurrentCommitIndex int
	GraphWidth         int
	SelectedBranch     Branch
	CurrentBranchName  string
	repo               *repo
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
	ID    string
	Name  string
	Index int
}

func newViewPort(repo *repo, first, last, selected int) ViewPort {
	if last >= len(repo.Commits) {
		last = len(repo.Commits) - 1
	}
	if first >= len(repo.Commits) {
		first = -1
		last = -1
	}

	return ViewPort{
		Commits:           toCommits(repo, first, last),
		FirstIndex:        first,
		LastIndex:         last,
		TotalCommits:      len(repo.Commits),
		CurrentBranchName: repo.CurrentBranchName,
		GraphWidth:        len(repo.Branches) * 2,
		SelectedBranch:    toBranch(repo.Commits[selected].Branch),
		repo:              repo,
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
		ID:    b.id,
		Name:  b.name,
		Index: b.index,
	}
}
