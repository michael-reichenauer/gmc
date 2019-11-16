package model

type branch struct {
	id            string
	name          string
	index         int
	tipId         string
	tip           *commit
	bottom        *commit
	parentBranch  *branch
	childBranches []*branch
	isGitBranch   bool
}

func (b *branch) String() string {
	return b.name
}

func (b *branch) isAncestor(dc *branch) bool {
	for dc.parentBranch != nil {
		if dc.parentBranch == b {
			return true
		}
		dc = dc.parentBranch
	}
	return false
}
