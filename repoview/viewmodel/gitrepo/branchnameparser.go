package gitrepo

import (
	"regexp"
)

var (
	nameRegExp = regexp.MustCompile( // parse subject like e.g. "Merge branch 'develop' into master"
		`[Mm]erged?` + //                                     'Merge' or 'merged' word
			`(\s+remote-tracking)?` + //                      'remote-tracking' optional word when merging remote branches
			`(\s+(from branch|branch|commit|from))?` + //                 'branch'|'commit'|'from' word
			`\s+'?(?P<from>[0-9A-Za-z_/-]+)'?` + //           the <from> branch name
			`(?P<direction>\s+of\s+[^\s]+)?` + //             the optional 'of repo url'
			`(\s+(into|to)\s+(?P<into>[0-9A-Za-z_/-]+))?`) // the <into> branch name
	from, into, direction = indexes()
)

type fromInto struct {
	from string
	into string
}

type branchNameParser struct {
	parsedCommits map[string]fromInto
	branchNames   map[string]string
}

func newBranchNameParser() *branchNameParser {
	return &branchNameParser{
		parsedCommits: make(map[string]fromInto),
		branchNames:   make(map[string]string),
	}
}

func (h *branchNameParser) parseCommit(c *Commit) fromInto {
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

func (h *branchNameParser) branchName(id string) string {
	return h.branchNames[id]
}

func (h *branchNameParser) parseMergeBranchNames(subject string) fromInto {
	matches := nameRegExp.FindAllStringSubmatch(subject, -1)
	if len(matches) == 0 {
		return fromInto{}
	}

	if matches[0][from] != "" && matches[0][direction] != "" && matches[0][into] == "" {
		// Subject is a pull merge (same source and target branch)
		return fromInto{from: matches[0][from], into: matches[0][from]}
	}
	return fromInto{from: matches[0][from], into: matches[0][into]}
}

// indexes returns the named group indexes to be used in parse
func indexes() (fromIndex, intoIndex, directionIndex int) {
	n1 := nameRegExp.SubexpNames()
	for i, v := range n1 {
		if v == "from" {
			fromIndex = i
		}
		if v == "into" {
			intoIndex = i
		}
		if v == "direction" {
			directionIndex = i
		}
	}
	return
}
