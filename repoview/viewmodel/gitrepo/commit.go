package gitrepo

import (
	"fmt"
	"github.com/michael-reichenauer/gmc/utils/git"
	"strings"
	"time"
)

type Commit struct {
	Id         string
	Sid        string
	Subject    string
	Message    string
	Author     string
	AuthorTime time.Time
	ParentIDs  []string

	IsCurrent      bool
	ChildIDs       []string
	FirstParent    *Commit
	MergeParent    *Commit
	Children       []*Commit
	MergeChildren  []*Commit
	Branch         *Branch
	Branches       []*Branch
	BranchTipNames []string
	isLikely       bool
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
	}
}

func newPartialLogCommit() *Commit {
	return &Commit{
		Id:         git.PartialLogCommitID,
		Sid:        git.ToSid(git.PartialLogCommitID),
		ParentIDs:  []string{},
		Subject:    "...    (more commits)",
		Message:    "...    (more commits)",
		Author:     "",
		AuthorTime: time.Date(2000, 1, 1, 1, 1, 0, 0, time.UTC),
	}
}
func (c *Commit) contains(lowerText string) bool {
	return contains(c.Id, lowerText) ||
		contains(c.Message, lowerText) ||
		contains(c.Author, lowerText)
}

func contains(text, lowerSearch string) bool {
	return strings.Contains(strings.ToLower(text), lowerSearch)
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
		if v != b[i] && v.Name != b[i].Name {
			return false
		}
	}
	return true
}
