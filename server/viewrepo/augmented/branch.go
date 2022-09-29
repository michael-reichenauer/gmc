package augmented

import (
	"fmt"

	"github.com/michael-reichenauer/gmc/utils/git"
)

type Branch struct {
	Name              string
	DisplayName       string
	TipID             string
	BottomID          string
	ParentBranch      *Branch
	IsRemote          bool
	RemoteName        string
	LocalName         string
	IsCurrent         bool
	IsGitBranch       bool
	IsAmbiguousBranch bool
	IsNamedBranch     bool
	IsSetAsParent     bool
	AmbiguousBranches []*Branch
}

func newGitBranch(gb git.Branch) *Branch {
	return &Branch{
		Name:        gb.Name,
		DisplayName: gb.DisplayName,
		TipID:       gb.TipID,
		IsCurrent:   gb.IsCurrent,
		IsRemote:    gb.IsRemote,
		RemoteName:  gb.RemoteName,
		IsGitBranch: true,
	}
}

func newNamedBranch(commitID string, branchName string) *Branch {
	sid := git.ToSid(commitID)
	return &Branch{
		Name:          fmt.Sprintf("%s:%s", branchName, sid),
		DisplayName:   branchName,
		TipID:         commitID,
		IsNamedBranch: true,
	}
}

func newBranch(commitID string) *Branch {
	sid := git.ToSid(commitID)
	return &Branch{
		Name:          fmt.Sprintf("branch:%s", sid),
		DisplayName:   fmt.Sprintf("branch@%s", sid),
		TipID:         commitID,
		IsNamedBranch: true,
	}
}

func newAmbiguousBranch(commitID string) *Branch {
	sid := git.ToSid(commitID)
	return &Branch{
		Name:              fmt.Sprintf("ambiguous:%s", sid),
		DisplayName:       fmt.Sprintf("ambiguous@%s", sid),
		TipID:             commitID,
		IsAmbiguousBranch: true,
	}
}

func (t *Branch) IsAncestorBranch(name string) bool {
	for b := t; b != nil && b.ParentBranch != nil; b = b.ParentBranch {
		if name == b.ParentBranch.Name {
			return true
		}
	}
	return false
}

func (t *Branch) GetAncestorsAndSelf() []*Branch {
	var branches []*Branch
	for b := t; b != nil; b = b.ParentBranch {
		branches = append(branches, b)
	}
	return branches
}

func (t *Branch) GetAncestors() []*Branch {
	var branches []*Branch
	for b := t.ParentBranch; b != nil; b = b.ParentBranch {
		branches = append(branches, b)
	}
	return branches
}

func (t *Branch) String() string {
	return t.Name
}
