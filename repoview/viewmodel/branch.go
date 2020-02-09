package viewmodel

type branch struct {
	name             string
	displayName      string
	index            int
	tipId            string
	bottomId         string
	parentBranchName string
	remoteName       string
	localName        string
	tip              *commit
	bottom           *commit
	parentBranch     *branch
	isGitBranch      bool
	isMultiBranch    bool
	isRemote         bool
	isCurrent        bool
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
