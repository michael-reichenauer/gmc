package gitmodel

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

func (b *Branch) String() string {
	return b.DisplayName
}
