package gitmodel

import (
	"fmt"
	"github.com/michael-reichenauer/gmc/utils/git"
)

type Branch struct {
	ID            string
	Name          string
	TipID         string
	IsCurrent     bool
	IsGitBranch   bool
	IsMultiBranch bool
	IsNamedBranch bool
}

func newBranch(gb git.Branch) *Branch {
	return &Branch{
		ID:          toBranchID(gb),
		Name:        gb.Name,
		TipID:       gb.TipID,
		IsCurrent:   gb.IsCurrent,
		IsGitBranch: true,
	}
}

func toBranchID(gb git.Branch) string {
	return fmt.Sprintf("%s", gb.ID)
}
func (b *Branch) String() string {
	return b.Name
}
