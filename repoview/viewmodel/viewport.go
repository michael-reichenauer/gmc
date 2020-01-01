package viewmodel

import (
	"time"
)

type ViewPort struct {
	Commits            []Commit
	FirstIndex         int
	TotalCommits       int
	GraphWidth         int
	CurrentBranchName  string
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
	IsCurrent  bool
	Branch     Branch
	Graph      []GraphColumn
	IsMore     bool
	ParentIDs  []string
	ChildIDs   []string
	BranchTips []string
}

type Branch struct {
	Name          string
	DisplayName   string
	Index         int
	IsMultiBranch bool
	RemoteName    string
	LocalName     string
	IsRemote      bool
	IsGitBranch   bool
}

func newViewPort(repo *repo, firstIndex, count int) ViewPort {
	return ViewPort{
		Commits:            toCommits(repo, firstIndex, count),
		FirstIndex:         firstIndex,
		TotalCommits:       len(repo.Commits),
		CurrentBranchName:  repo.CurrentBranchName,
		GraphWidth:         len(repo.Branches) * 2,
		RepoPath:           repo.gmRepo.RepoPath,
		UncommittedChanges: repo.gmStatus.AllChanges(),
	}
}

func toCommits(repo *repo, firstIndex int, count int) []Commit {
	commits := make([]Commit, count)
	for i := 0; i < count; i++ {
		commits[i] = toCommit(repo.Commits[i+firstIndex])
	}
	return commits
}

func toCommit(c *commit) Commit {
	return Commit{
		ID:         c.ID,
		SID:        c.SID,
		Subject:    c.Subject,
		Message:    c.Message,
		ParentIDs:  c.ParentIDs,
		ChildIDs:   c.ChildIDs,
		Author:     c.Author,
		AuthorTime: c.AuthorTime,
		IsCurrent:  c.IsCurrent,
		Branch:     toBranch(c.Branch),
		Graph:      c.graph,
		IsMore:     c.IsMore,
		BranchTips: c.BranchTips,
	}
}

func toBranch(b *branch) Branch {
	return Branch{
		Name:          b.name,
		DisplayName:   b.displayName,
		Index:         b.index,
		IsMultiBranch: b.isMultiBranch,
		RemoteName:    b.remoteName,
		LocalName:     b.localName,
		IsRemote:      b.isRemote,
		IsGitBranch:   b.isGitBranch,
	}
}
