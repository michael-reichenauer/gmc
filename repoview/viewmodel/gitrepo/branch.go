package gitrepo

import (
	"github.com/michael-reichenauer/gmc/utils/git"
)

type Branch struct {
	Name          string
	DisplayName   string
	TipID         string
	BottomID      string
	ParentBranch  *Branch
	IsRemote      bool
	RemoteName    string
	LocalName     string
	IsCurrent     bool
	IsGitBranch   bool
	IsMultiBranch bool
	IsNamedBranch bool
}

func newBranch(gb git.Branch) *Branch {
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
