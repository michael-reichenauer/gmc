package gitmodel

import (
	"fmt"
	"github.com/michael-reichenauer/gmc/utils/git"
	"time"
)

type Commit struct {
	Id            string
	Sid           string
	Subject       string
	Message       string
	Author        string
	AuthorTime    time.Time
	IsCurrent     bool
	ParentIDs     []string
	ChildIDs      []string
	Parent        *Commit
	MergeParent   *Commit
	Children      []*Commit
	MergeChildren []*Commit
	Branch        *Branch
	Branches      []*Branch
}

func newCommit(gc git.Commit) *Commit {
	return &Commit{
		Id:         gc.ID,
		Sid:        gc.SID,
		ParentIDs:  gc.ParentIDs,
		Subject:    gc.Subject,
		Message:    gc.Message,
		Author:     gc.Author,
		AuthorTime: gc.AuthorTime,
		IsCurrent:  false,
	}
}

func (c *Commit) addBranches(branches []*Branch) {
	if c.Branches == nil {
		c.Branches = branches
		return
	}
	if areEqual(c.Branches, branches) {
		return
	}
	for _, b := range branches {
		c.addBranch(b)
	}
}

func (c *Commit) addBranch(branch *Branch) {
	if c.hasBranch(branch) {
		return
	}
	c.Branches = append(c.Branches, branch)
}

func (c *Commit) hasBranch(branch *Branch) bool {
	for _, b := range c.Branches {
		if b == branch {
			return true
		}
	}
	return false
}

func (c *Commit) String() string {
	return fmt.Sprintf("%s %s (%s)", c.Sid, c.Subject, c.Branch)
}

func areEqual(a, b []*Branch) bool {
	if len(a) != len(b) {
		return false
	}
	for i, v := range a {
		if v != b[i] && v.ID != b[i].ID {
			return false
		}
	}
	return true
}
