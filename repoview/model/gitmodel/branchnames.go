package gitmodel

import "strings"

func parseBranchNames(c *Commit) (from, into string) {
	if len(c.ParentIDs) == 2 {
		return parseMergeBranchNames(c.Subject)
	}
	return "", ""
}

func parseMergeBranchNames(subject string) (from string, into string) {
	if strings.HasPrefix(subject, "Merge branch '") {
		ei := strings.LastIndex(subject, "'")
		if ei > 14 {
			from = subject[14:ei]
			if strings.HasPrefix(subject[ei:], "' into ") {
				into = subject[ei+7:]
			}
		}
	}
	return
}
