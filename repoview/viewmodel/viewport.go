package viewmodel

import (
	"github.com/michael-reichenauer/gmc/utils"
	"time"
)

type ViewRepo struct {
	Commits            []Commit
	CurrentBranchName  string
	RepoPath           string
	UncommittedChanges int
	viewRepo           *viewRepo
}

const (
	MoreNone                    = 0
	MoreMergeIn   utils.Bitmask = 1 << iota // ╮
	MoreBranchOut                           // ╭
)

type Commit struct {
	ID           string
	SID          string
	Subject      string
	Message      string
	Author       string
	AuthorTime   time.Time
	IsCurrent    bool
	Branch       Branch
	Graph        []GraphColumn
	More         utils.Bitmask
	ParentIDs    []string
	ChildIDs     []string
	BranchTips   []string
	IsLocalOnly  bool
	IsRemoteOnly bool
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
	IsCurrent     bool
	TipID         string
	HasLocalOnly  bool
	HasRemoteOnly bool
}

type Branches []Branch

func (bs Branches) Contains(predicate func(b Branch) bool) bool {
	for _, b := range bs {
		if predicate(b) {
			return true
		}
	}
	return false
}

func newViewRepo(repo *viewRepo) ViewRepo {
	return ViewRepo{
		Commits:            toCommits(repo),
		CurrentBranchName:  repo.CurrentBranchName,
		RepoPath:           repo.WorkingFolder,
		UncommittedChanges: repo.UncommittedChanges,
		viewRepo:           repo,
	}
}

func toCommits(repo *viewRepo) []Commit {
	commits := make([]Commit, len(repo.Commits))
	for i, c := range repo.Commits {
		commits[i] = toCommit(c)
	}
	return commits
}

func toCommit(c *commit) Commit {
	return Commit{
		ID:           c.ID,
		SID:          c.SID,
		Subject:      c.Subject,
		Message:      c.Message,
		ParentIDs:    c.ParentIDs,
		ChildIDs:     c.ChildIDs,
		Author:       c.Author,
		AuthorTime:   c.AuthorTime,
		IsCurrent:    c.IsCurrent,
		Branch:       toBranch(c.Branch),
		Graph:        c.graph,
		More:         c.More,
		BranchTips:   c.BranchTips,
		IsLocalOnly:  c.IsLocalOnly,
		IsRemoteOnly: c.IsRemoteOnly,
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
		TipID:         b.tipId,
		IsCurrent:     b.isCurrent,
		HasRemoteOnly: b.HasRemoteOnly,
		HasLocalOnly:  b.HasLocalOnly,
	}
}
