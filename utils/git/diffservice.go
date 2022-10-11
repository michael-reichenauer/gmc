	"github.com/samber/lo"
	Id        string
	Author    string
	Date      string
	Message   string
	commitDiffs, err := t.parse(diffText, "", false)
	return commitDiffs[0], nil
}

func (t *diffService) fileDiff(path string) ([]CommitDiff, error) {
	diffText, err := t.cmd.Git("log", "--patch", "--follow", "--", path)
	if err != nil {
		return []CommitDiff{}, err
	}
	commitDiffs, err := t.parse(diffText, path, false)
	if err != nil {
		return []CommitDiff{}, err
	}
	return commitDiffs, nil
	commitDiffs, err := t.parse(diffText, "", true)
	fileDiffs := commitDiffs[0].FileDiffs
		return strings.Compare(strings.ToLower(fileDiffs[i].PathAfter), strings.ToLower(fileDiffs[j].PathAfter)) == -1
func (t *diffService) parse(text, path string, isUncommitted bool) ([]CommitDiff, error) {
	var currentCommitDiff *CommitDiff
	var commitDiffs []CommitDiff

	if isUncommitted {
		// Uncommitted diffs do not have a commit header
		commitDiffs = append(commitDiffs, CommitDiff{Id: UncommittedID, FileDiffs: []FileDiff{}})
		currentCommitDiff = &commitDiffs[len(commitDiffs)-1]
	}

	for i, line := range lines {
		if cd, ok := tryParseCommitDiffHead(i, lines); ok {
			commitDiffs = append(commitDiffs, cd)
			currentCommitDiff = &commitDiffs[len(commitDiffs)-1]
			currentFileDiff = nil
			continue
		}
			if path != "" && fd.PathBefore != path && fd.PathAfter != path {
				// File history includes all paths where path is within the file paths,
				// Lets filter away those that do not match
				currentFileDiff = &fd
				currentSectionDiff = nil
				continue
			}
			currentCommitDiff.FileDiffs = append(currentCommitDiff.FileDiffs, fd)
			currentFileDiff = &currentCommitDiff.FileDiffs[len(currentCommitDiff.FileDiffs)-1]

	// For file history with other similar file paths, the file diff can be empty, lets filter them
	commitDiffs = lo.Filter(commitDiffs, func(v CommitDiff, _ int) bool {
		return len(v.FileDiffs) > 0
	})

	return commitDiffs, nil
func tryParseCommitDiffHead(index int, lines []string) (CommitDiff, bool) {
	line := lines[index]

	author := ""
	date := ""
	if !strings.HasPrefix(line, "commit ") {
		return CommitDiff{}, false
	}
	if len(lines) > index+2 && strings.HasPrefix(lines[index+1], "Author: ") {
		author = lines[index+1][len("Author: "):]
	}
	if len(lines) > index+3 && strings.HasPrefix(lines[index+2], "Date:   ") {
		date = lines[index+2][len("Date:   "):]
	}
	message := lines[index+4]
	message = strings.TrimSpace(message)

	commitId := line[len("commit "):]
	return CommitDiff{Id: commitId, Author: author, Date: date, Message: message, FileDiffs: []FileDiff{}}, true
}

	before := parts[0][2:]
	after := parts[1][2:]
	fileDiff.DiffMode = DiffModified
	fileDiff.PathBefore = before
	fileDiff.PathAfter = after
