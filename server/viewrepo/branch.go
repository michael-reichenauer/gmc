package viewrepo

import (
	"github.com/michael-reichenauer/gmc/utils/cui"
)

type branch struct {
	index                int
	name                 string
	displayName          string
	tipId                string
	bottomId             string
	parentBranchName     string
	remoteName           string
	localName            string
	tip                  *commit
	bottom               *commit
	parentBranch         *branch
	isGitBranch          bool
	isAmbiguousBranch    bool
	isRemote             bool
	isCurrent            bool
	isSetAsParent        bool
	HasLocalOnly         bool
	HasRemoteOnly        bool
	color                cui.Color
	AmbiguousTipId       string
	ambiguousBranchNames []string
	x                    int
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
