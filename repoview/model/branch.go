package model

type branch struct {
	name           string
	displayName    string
	index          int
	tipId          string
	bottomId       string
	parentBranchID string
	tip            *commit
	bottom         *commit
	parentBranch   *branch
	isGitBranch    bool
	isMultiBranch  bool
}

func (b *branch) String() string {
	return b.displayName
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
