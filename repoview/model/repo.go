package model

import (
	"fmt"
	"gmc/repoview/model/gitmodel"
	"gmc/utils/log"
)

type repo struct {
	Commits           []*commit
	commitById        map[string]*commit
	Branches          []*branch
	CurrentCommit     *commit
	CurrentBranchName string
	gitRepo           gitmodel.Repo
}

func newRepo() *repo {
	return &repo{
		commitById: make(map[string]*commit),
	}
}

func (r *repo) CommitById(id string) (*commit, bool) {
	c, ok := r.commitById[id]
	return c, ok
}

func (r *repo) BranchById(id string) *branch {
	for _, b := range r.Branches {
		if id == b.id {
			return b
		}
	}
	panic("unknown branch id" + id)
}

func (r *repo) addBranch(gb *gitmodel.Branch) {
	b := r.toBranch(gb, len(r.Branches))
	r.Branches = append(r.Branches, b)
	if gb.IsCurrent {
		r.CurrentBranchName = gb.Name
	}
}

func (r *repo) addCommit(gmc *gitmodel.Commit) {
	if !r.containsBranch(gmc.Branch) {
		return
	}
	c := r.toCommit(gmc, len(r.Commits))
	r.Commits = append(r.Commits, c)
	r.commitById[c.ID] = c
	if c.IsCurrent {
		r.CurrentCommit = c
	}
}

func (r *repo) containsOneOfBranches(branches []*gitmodel.Branch) bool {
	for _, rb := range r.Branches {
		for _, b := range branches {
			if rb.id == b.ID {
				return true
			}
		}
	}
	return false
}

func (r *repo) containsBranch(branch *gitmodel.Branch) bool {
	for _, b := range r.Branches {
		if b.id == branch.ID {
			return true
		}
	}
	return false
}

func (r *repo) toBranch(b *gitmodel.Branch, index int) *branch {
	return &branch{
		id:          b.ID,
		name:        b.Name,
		index:       index,
		tipId:       b.TipID,
		isGitBranch: b.IsGitBranch,
	}

}

func (r *repo) toCommit(c *gitmodel.Commit, index int) *commit {
	var branch *branch
	if c.Branch != nil {
		branch = r.BranchById(c.Branch.ID)
	}
	return &commit{
		ID:         c.Id,
		SID:        c.Sid,
		Subject:    c.Subject,
		Message:    c.Message,
		Author:     c.Author,
		AuthorTime: c.AuthorTime,
		ParentIDs:  c.ParentIDs,
		ChildIDs:   c.ChildIDs,
		IsCurrent:  c.IsCurrent,
		Branch:     branch,
		Index:      index,
		graph:      make([]GraphColumn, len(r.Branches)),
	}
}

func (r *repo) String() string {
	return fmt.Sprintf("b:%d c:%d", len(r.Branches), len(r.Commits))
}

func (r *repo) ToBranchIndex(id string) int {
	for i, b := range r.Branches {
		if b.id == id {
			return i
		}
	}
	log.Fatalf("unexpected branch %s", id)
	return 0
}
