package gitmodel

import (
	"strings"
)

type fromInto struct {
	from string
	into string
}

type branchNames struct {
	parsedCommits map[string]fromInto
	branchNames   map[string]string
}

func newBranchNames() *branchNames {
	return &branchNames{
		parsedCommits: make(map[string]fromInto),
		branchNames:   make(map[string]string),
	}
}

func (h *branchNames) parseCommit(c *Commit) fromInto {
	if len(c.ParentIDs) != 2 {
		return fromInto{}
	}
	if fi, ok := h.parsedCommits[c.Id]; ok {
		// Already parsed this commit
		return fi
	}

	fi := h.parseMergeBranchNames(c.Subject)

	// set the branch name of the commit and merge parent.
	// could actually be multiple names, but lets ignore that
	h.parsedCommits[c.Id] = fi
	h.branchNames[c.Id] = fi.into
	h.branchNames[c.ParentIDs[1]] = fi.from
	return fi
}

func (h *branchNames) branchName(id string) string {
	return h.branchNames[id]
}

func (h *branchNames) parseMergeBranchNames(subject string) fromInto {
	var fi fromInto
	if strings.HasPrefix(subject, "Merge branch '") {
		ei := strings.LastIndex(subject, "'")
		if ei > 14 {
			fi.from = subject[14:ei]
			if strings.HasPrefix(subject[ei:], "' into ") {
				fi.into = subject[ei+7:]
			}
		}
	}
	return fi
}
