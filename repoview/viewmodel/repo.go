package viewmodel

import (
	"fmt"
	"github.com/michael-reichenauer/gmc/repoview/viewmodel/gitrepo"
	"github.com/michael-reichenauer/gmc/utils/git"
	"github.com/michael-reichenauer/gmc/utils/log"
	"time"
)

const (
	StatusID  = git.UncommittedID
	StatusSID = git.UncommittedSID
)

type viewRepo struct {
	Commits            []*commit
	commitById         map[string]*commit
	Branches           []*branch
	CurrentCommit      *commit
	CurrentBranchName  string
	WorkingFolder      string
	UncommittedChanges int
	gitRepo            gitrepo.Repo
}

func newRepo() *viewRepo {
	return &viewRepo{
		commitById: make(map[string]*commit),
	}
}

func (r *viewRepo) CommitById(id string) (*commit, bool) {
	c, ok := r.commitById[id]
	return c, ok
}

func (r *viewRepo) BranchByName(name string) *branch {
	for _, b := range r.Branches {
		if name == b.name {
			return b
		}
	}
	panic(log.Fatal(fmt.Errorf("unknown branch id %s", name)))
}

func (r *viewRepo) addBranch(gb *gitrepo.Branch) {
	b := r.toBranch(gb, len(r.Branches))
	r.Branches = append(r.Branches, b)
}

func (r *viewRepo) addVirtualStatusCommit(gRepo gitrepo.Repo) {
	if gRepo.Status.OK() {
		return
	}
	cb, ok := gRepo.CurrentBranch()
	if !ok || !r.containsBranch(cb) {
		return
	}
	allChanges := gRepo.Status.AllChanges()
	statusText := fmt.Sprintf("%d uncommitted changes", allChanges)
	if gRepo.Status.IsMerging && gRepo.Status.MergeMessage != "" {
		statusText = statusText + ", " + gRepo.Status.MergeMessage
	}

	c := r.toVirtualStatusCommit(cb.Name, statusText, len(r.Commits))
	r.Commits = append(r.Commits, c)
	r.commitById[c.ID] = c
}

func (r *viewRepo) addGitCommit(gc *gitrepo.Commit) {
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

func (r *viewRepo) containsOneOfBranches(branches []*gitrepo.Branch) bool {
	for _, rb := range r.Branches {
		for _, b := range branches {
			if rb.name == b.Name {
				return true
			}
		}
	}
	return false
}

func (r *viewRepo) containsBranch(branch *gitrepo.Branch) bool {
	for _, b := range r.Branches {
		if b.name == branch.Name {
			return true
		}
	}
	return false
}

func (r *viewRepo) containsBranchName(name string) bool {
	for _, b := range r.Branches {
		if b.name == name {
			return true
		}
	}
	return false
}

func (r *viewRepo) toBranch(b *gitrepo.Branch, index int) *branch {
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
		isCurrent:        b.IsCurrent,
	}
}

func (r *viewRepo) toCommit(c *gitrepo.Commit, index int) *commit {
	var branch = r.BranchByName(c.Branch.Name)
	isLocalOnly := r.isLocalOnly(c, branch)
	isRemoteOnly := r.isRemoteOnly(c, branch)

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
		BranchTips:   c.BranchTipNames,
		IsLocalOnly:  isLocalOnly,
		IsRemoteOnly: isRemoteOnly,
	}
}

func (r *viewRepo) isRemoteOnly(c *gitrepo.Commit, branch *branch) bool {
	if !c.Branch.IsGitBranch || !c.Branch.IsRemote || c.Branch.LocalName == "" {
		// Only remote git branches, with local branch can have remote only remote commits
		return false
	}
	// the commit branch is remote and a local branch exist as well
	if r.containsGitBranchName(c.Branches, c.Branch.LocalName) {
		// The commit branches does contain the local branch and thus this commit is pushed
		return false
	}

	// But the remote branch may have been pulled/merged into the local, lets iterate all
	// commits in the remote branch and check if they have been merged into the local
	tip := r.commitById[branch.tipId]
	for current := tip; current != nil && len(current.ParentIDs) > 0; current = r.commitById[current.ParentIDs[0]] {
		for _, cid := range current.ChildIDs {
			child := r.commitById[cid]
			if child != nil && child.Branch.name == c.Branch.LocalName {
				// the commit was merged/pulled into the local branch, thus commit has been pulled
				return false
			}
		}
	}

	// Since the current commit has not yet been added to the repo, the above iteration did
	// not cover the current commit, lets do that manually
	for _, mc := range c.MergeChildren {
		if mc.Branch.Name == c.Branch.LocalName {
			// the commit was merged/pulled into the local branch, thus commit has been pulled
			return false
		}
	}
	return true
}

func (r *viewRepo) isLocalOnly(c *gitrepo.Commit, branch *branch) bool {
	if !c.Branch.IsGitBranch || c.Branch.IsRemote || c.Branch.RemoteName == "" {
		// Only local git branches, with remote branch can have local only remote commits
		return false
	}
	// Commit branch is a local git branch and a remote git branch exist as well
	if r.containsGitBranchName(c.Branches, c.Branch.RemoteName) {
		// The commit branches does contain the remote branch and thus this commit has been pushed
		return false
	}

	// But the local branch may have been pulled/merged into the remote, lets iterate all
	// commits in the local branch and check if they have been merged into the remote
	tip := r.commitById[branch.tipId]
	for current := tip; current != nil && len(current.ParentIDs) > 0; current = r.commitById[current.ParentIDs[0]] {
		for _, cid := range current.ChildIDs {
			child := r.commitById[cid]
			if child != nil && child.Branch.name == c.Branch.RemoteName {
				// the commit was merged into the remote branch, thus commit has been pushed
				return false
			}
		}
	}

	// Since the current commit has not yet been added to the repo, the above iteration did
	// not cover the current commit, lets do that manually
	for _, mc := range c.MergeChildren {
		if mc.Branch.Name == c.Branch.RemoteName {
			// the commit was merged into the remote branch, thus commit has been pushed
			return false
		}
	}

	return true
}

func (r *viewRepo) containsGitBranchName(branches []*gitrepo.Branch, name string) bool {
	for _, b := range branches {
		if name == b.Name {
			return true
		}
	}
	return false
}

func (r *viewRepo) toVirtualStatusCommit(branchName string, statusText string, index int) *commit {
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

func (r *viewRepo) String() string {
	return fmt.Sprintf("b:%d c:%d", len(r.Branches), len(r.Commits))
}

func (r *viewRepo) ToBranchIndex(id string) int {
	for i, b := range r.Branches {
		if b.name == id {
			return i
		}
	}

	panic(log.Fatal(fmt.Errorf("unexpected branch %s", id)))
}
