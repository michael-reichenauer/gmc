package gitrepo

import (
	"regexp"
	"strings"
)

var (
	prefixes   = []string{"refs/remotes/origin/", "remotes/origin/", "origin/"}
	nameRegExp = regexp.MustCompile( // parse subject like e.g. "Merge branch 'develop' into main"
		`[Mm]erged?` + //                                     'Merge' or 'merged' word
			`(\s+remote-tracking)?` + //                      'remote-tracking' optional word when merging remote branches
			`(\s+(from branch|branch|commit|from))?` + //                 'branch'|'commit'|'from' word
			`\s+'?(?P<from>[0-9A-Za-z_/-]+)'?` + //           the <from> branch name
			`(?P<direction>\s+of\s+[^\s]+)?` + //             the optional 'of repo url'
			`(\s+(into|to)\s+(?P<into>[0-9A-Za-z_/-]+))?`) // the <into> branch name
	from, into, direction = nameRegExpIndexes()
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

func (h *branchNameParser) isPullMerge(c *Commit) bool {
	fi := h.parseCommit(c)
	return h.isPullMergeCommit(fi)
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
	if !h.isPullMergeCommit(fi) {
		h.branchNames[c.ParentIDs[1]] = fi.from
	} else {
		// The order of the parents will be switched for a pull merge and thus
		h.branchNames[c.ParentIDs[0]] = fi.from
	}

	return fi
}

func (h *branchNameParser) isPullMergeCommit(fi fromInto) bool {
	return fi.from != "" && fi.from == fi.into
}

func (h *branchNameParser) branchName(id string) string {
	return h.branchNames[id]
}

func (h *branchNameParser) parseMergeBranchNames(subject string) fromInto {
	subject = strings.TrimSpace(subject)
	matches := nameRegExp.FindAllStringSubmatch(subject, -1)
	if len(matches) == 0 {
		return fromInto{}
	}
	match := matches[0]

	if h.isMatchPullMerge(match) {
		// Subject is a pull merge same branch from remote repo (same remote source and target branch)
		return fromInto{
			from: h.trimBranchName(match[from]),
			into: h.trimBranchName(match[from])}
	}

	return fromInto{
		from: h.trimBranchName(match[from]),
		into: h.trimBranchName(match[into])}
}

func (h *branchNameParser) isMatchPullMerge(match []string) bool {
	if match[from] != "" && match[direction] != "" &&
		(match[into] == "" || match[into] == match[from]) {
		return true
	}

	if match[from] != "" && match[into] != "" &&
		h.trimBranchName(match[from]) == h.trimBranchName(match[into]) {
		return true
	}

	return false
}

func (h *branchNameParser) trimBranchName(name string) string {
	for _, prefix := range prefixes {
		if strings.HasPrefix(name, prefix) {
			return name[len(prefix):]
		}
	}
	return name
}

// nameRegExpIndexes returns the named group indexes to be used in parse
func nameRegExpIndexes() (fromIndex, intoIndex, directionIndex int) {
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
