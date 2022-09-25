package viewrepo

import (
	"fmt"
	"time"

	"github.com/michael-reichenauer/gmc/api"
	"github.com/michael-reichenauer/gmc/server/viewrepo/gitrepo"
	"github.com/michael-reichenauer/gmc/utils/git"
	"github.com/michael-reichenauer/gmc/utils/log"
)

type repo struct {
	Commits            []*commit
	commitById         map[string]*commit
	Branches           []*branch
	CurrentCommit      *commit
	CurrentBranchName  string
	WorkingFolder      string
	UncommittedChanges int
	gitRepo            gitrepo.Repo
	Conflicts          int
	MergeMessage       string
}

func newRepo() *repo {
	return &repo{
		commitById: make(map[string]*commit),
	}
}

func (t *repo) CommitById(id string) (*commit, bool) {
	c, ok := t.commitById[id]
	return c, ok
}

func (t *repo) BranchByName(name string) *branch {
	b := t.tryGetBranchByName(name)
	if b == nil {
		panic(log.Fatal(fmt.Errorf("unknown branch id %s", name)))
	}
	return b
}

func (t *repo) tryGetBranchByName(name string) *branch {
	for _, b := range t.Branches {
		if name == b.name {
			return b
		}
	}
	return nil
}

func (t *repo) addBranch(gb *gitrepo.Branch) {
	b := t.toBranch(gb, len(t.Branches))
	t.Branches = append(t.Branches, b)
}

func (t *repo) addVirtualStatusCommit(gRepo gitrepo.Repo) {
	cb, ok := gRepo.CurrentBranch()
	if !ok || !t.containsBranch(cb) {
		// No current branch, or view repo does not show the current branch
		return
	}

	allChanges := gRepo.Status.AllChanges()
	statusText := fmt.Sprintf("%d uncommitted changes", allChanges)
	if gRepo.Status.IsMerging && gRepo.Status.MergeMessage != "" {
		statusText = fmt.Sprintf("%s, %s", gRepo.Status.MergeMessage, statusText)
	}
	if gRepo.Status.Conflicted > 0 {
		statusText = fmt.Sprintf("CONFLICTS: %d, %s", gRepo.Status.Conflicted, statusText)
	}

	c := t.toVirtualStatusCommit(cb.Name, statusText, len(t.Commits))
	t.Commits = append(t.Commits, c)
	t.commitById[c.ID] = c
}

func (t *repo) addSearchCommit(gc *gitrepo.Commit) {
	c := t.toCommit(gc, len(t.Commits), false)
	t.Commits = append(t.Commits, c)
	t.commitById[c.ID] = c
	if c.IsCurrent {
		t.CurrentCommit = c
	}
}

func (t *repo) addGitCommit(gc *gitrepo.Commit) {
	if !t.containsBranch(gc.Branch) {
		return
	}
	c := t.toCommit(gc, len(t.Commits), true)
	t.Commits = append(t.Commits, c)
	t.commitById[c.ID] = c
	if c.IsCurrent {
		t.CurrentCommit = c
	}
}

func (t *repo) containsOneOfBranches(branches []*gitrepo.Branch) bool {
	for _, rb := range t.Branches {
		for _, b := range branches {
			if rb.name == b.Name {
				return true
			}
		}
	}
	return false
}

func (t *repo) containsBranch(branch *gitrepo.Branch) bool {
	for _, b := range t.Branches {
		if b.name == branch.Name {
			return true
		}
	}
	return false
}

func (t *repo) containsBranchName(name string) bool {
	for _, b := range t.Branches {
		if b.name == name {
			return true
		}
	}
	return false
}

func (t *repo) toBranch(b *gitrepo.Branch, index int) *branch {
	parentBranchName := ""
	if b.ParentBranch != nil {
		parentBranchName = b.ParentBranch.Name
	}
	var multiBranchNames []string
	for _, bb := range b.MultiBranches {
		multiBranchNames = append(multiBranchNames, bb.Name)
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
		isCurrent:        b.IsCurrent,
		isSetAsParent:    b.IsSetAsParent,
		multiBranchNames: multiBranchNames,
	}
}

func (t *repo) toCommit(c *gitrepo.Commit, index int, includeGraph bool) *commit {
	var branch = t.BranchByName(c.Branch.Name)

	var graph []api.GraphColumn
	if includeGraph {
		graph = make([]api.GraphColumn, len(t.Branches))
	}
	var multiBranches []string
	if branch.isMultiBranch {
		for _, b := range c.Branches {
			multiBranches = append(multiBranches, b.Name)
		}
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
		graph:      graph,
		BranchTips: c.BranchTipNames,
	}
}

func (t *repo) containsGitBranchName(branches []*gitrepo.Branch, name string) bool {
	for _, b := range branches {
		if name == b.Name {
			return true
		}
	}
	return false
}

func (t *repo) toVirtualStatusCommit(branchName string, statusText string, index int) *commit {
	branch := t.BranchByName(branchName)
	return &commit{
		ID:         git.UncommittedID,
		SID:        git.UncommittedSID,
		Subject:    statusText,
		Message:    statusText,
		Author:     "",
		AuthorTime: time.Now(),
		ParentIDs:  []string{branch.tipId},
		ChildIDs:   []string{},
		IsCurrent:  false,
		Branch:     branch,
		Index:      index,
		graph:      make([]api.GraphColumn, len(t.Branches)),
	}
}

func (t *repo) String() string {
	return fmt.Sprintf("b:%d c:%d", len(t.Branches), len(t.Commits))
}

func (t *repo) ToBranchIndex(id string) int {
	for i, b := range t.Branches {
		if b.name == id {
			return i
		}
	}

	panic(log.Fatal(fmt.Errorf("unexpected branch %s", id)))
}
