package viewrepo

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
	HasLocalOnly     bool
	HasRemoteOnly    bool
}

func (t *branch) String() string {
	return t.name
}

func (t *branch) isAncestor(dc *branch) bool {
	for dc.parentBranch != nil {
		if dc.parentBranch == t {
			return true
		}
		dc = dc.parentBranch
	}
	return false
}
