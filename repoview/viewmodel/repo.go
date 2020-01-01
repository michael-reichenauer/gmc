package viewmodel

import (
	"fmt"
	"github.com/michael-reichenauer/gmc/repoview/viewmodel/gitrepo"
	"github.com/michael-reichenauer/gmc/utils/log"
	"time"
)

const (
	StatusID  = "0000000000000000000000000000000000000000"
	StatusSID = "000000"
)

type repo struct {
	Commits           []*commit
	commitById        map[string]*commit
	Branches          []*branch
	CurrentCommit     *commit
	CurrentBranchName string
	gmRepo            gitrepo.Repo
	gmStatus          gitrepo.Status
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

func (r *repo) BranchByName(name string) *branch {
	for _, b := range r.Branches {
		if name == b.name {
			return b
		}
	}
	log.Fatal("unknown branch id" + name)
	return nil
}

func (r *repo) addBranch(gb *gitrepo.Branch) {
	b := r.toBranch(gb, len(r.Branches))
	r.Branches = append(r.Branches, b)
}

func (r *repo) addVirtualStatusCommit() {
	if r.gmStatus.OK() {
		return
	}
	cb, ok := r.gmRepo.CurrentBranch()
	if !ok || !r.containsBranch(cb) {
		return
	}
	allChanges := r.gmStatus.AllChanges()
	statusText := fmt.Sprintf("%d uncommitted changes", allChanges)

	c := r.toVirtualStatusCommit(cb.Name, statusText, len(r.Commits))
	r.Commits = append(r.Commits, c)
	r.commitById[c.ID] = c
}

func (r *repo) addGitCommit(gc *gitrepo.Commit) {
	if !r.containsBranch(gc.Branch) {
		return
	}
	c := r.toCommit(gc, len(r.Commits))
	r.Commits = append(r.Commits, c)
	r.commitById[c.ID] = c
	if c.IsCurrent {
		r.CurrentCommit = c
	}
}

func (r *repo) containsOneOfBranches(branches []*gitrepo.Branch) bool {
	for _, rb := range r.Branches {
		for _, b := range branches {
			if rb.name == b.Name {
				return true
			}
		}
	}
	return false
}

func (r *repo) containsBranch(branch *gitrepo.Branch) bool {
	for _, b := range r.Branches {
		if b.name == branch.Name {
			return true
		}
	}
	return false
}

func (r *repo) containsBranchName(name string) bool {
	for _, b := range r.Branches {
		if b.name == name {
			return true
		}
	}
	return false
}

func (r *repo) toBranch(b *gitrepo.Branch, index int) *branch {
	parentBranchName := ""
	if b.ParentBranch != nil {
		parentBranchName = b.ParentBranch.Name
	}
	return &branch{
		name:             b.Name,
		displayName:      b.DisplayName,
		index:            index,
		tipId:            b.TipID,
		bottomId:         b.BottomID,
		parentBranchName: parentBranchName,
		isGitBranch:      b.IsGitBranch,
		isRemote:         b.IsRemote,
		isMultiBranch:    b.IsMultiBranch,
		remoteName:       b.RemoteName,
		localName:        b.LocalName,
	}
}

func (r *repo) toCommit(c *gitrepo.Commit, index int) *commit {
	var branch *branch
	if c.Branch != nil {
		branch = r.BranchByName(c.Branch.Name)
	}
	isLocalOnly := false
	isRemoteOnly := false
	if c.Branch.IsGitBranch {
		if c.Branch.IsRemote && c.Branch.LocalName != "" {
			if !containsBranchName(c.Branches, c.Branch.LocalName) {
				isRemoteOnly = true
				log.Infof("Commit remote only %s", c)
			}
		}
		if !c.Branch.IsRemote && c.Branch.RemoteName != "" {
			if !containsBranchName(c.Branches, c.Branch.RemoteName) {
				isLocalOnly = true
				log.Infof("Commit local only %s", c)
			}
		}
	}

	return &commit{
		ID:           c.Id,
		SID:          c.Sid,
		Subject:      c.Subject,
		Message:      c.Message,
		Author:       c.Author,
		AuthorTime:   c.AuthorTime,
		ParentIDs:    c.ParentIDs,
		ChildIDs:     c.ChildIDs,
		IsCurrent:    c.IsCurrent,
		Branch:       branch,
		Index:        index,
		graph:        make([]GraphColumn, len(r.Branches)),
		BranchTips:   c.BranchTips,
		IsLocalOnly:  isLocalOnly,
		IsRemoteOnly: isRemoteOnly,
	}
}

func containsBranchName(branches []*gitrepo.Branch, name string) bool {
	for _, b := range branches {
		if name == b.Name {
			return true
		}
	}
	return false
}

func (r *repo) toVirtualStatusCommit(branchName string, statusText string, index int) *commit {
	branch := r.BranchByName(branchName)
	return &commit{
		ID:         StatusID,
		SID:        StatusSID,
		Subject:    statusText,
		Message:    statusText,
		Author:     "",
		AuthorTime: time.Now(),
		ParentIDs:  []string{branch.tipId},
		ChildIDs:   []string{},
		IsCurrent:  false,
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
		if b.name == id {
			return i
		}
	}
	log.Fatalf("unexpected branch %s", id)
	return 0
}
