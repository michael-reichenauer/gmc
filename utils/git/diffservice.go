	DiffConflicts
	DiffConflictStart
	DiffConflictSplit
	DiffConflictEnd
		"--find-renames", "--unified=6", "HEAD")
	t.setConflictsFilesMode(fileDiffs, status)
	fileDiffs, err = t.addAddedFiles(fileDiffs, status, t.cmd.WorkingDir())
		if fileDiffMode, ok := tryParseFileMode(line); ok && currentFileDiff != nil && currentFileDiff.DiffMode != DiffConflicts {
func (t *diffService) setConflictsFilesMode(diffs []FileDiff, status Status) {
	for i, fd := range diffs {
		if utils.StringsContains(status.ConflictsFiles, fd.PathAfter) {
			fd.DiffMode = DiffConflicts
			diffs[i] = fd
		}
	}
}

	if strings.HasPrefix(line, "diff --cc ") {
		return tryParseFileDiffConflictHead(line)
	}
func tryParseFileDiffConflictHead(line string) (FileDiff, bool) {
	var fileDiff FileDiff
	fileDiff.DiffMode = DiffConflicts
	file := line[10:]
	fileDiff.PathBefore = file
	fileDiff.PathAfter = file
	fileDiff.IsRenamed = false
	return fileDiff, true
}

	switch {
	case strings.HasPrefix(line, "><<<<<<<"):
		return LinesDiff{DiffMode: DiffConflictStart, Line: asConflictLine(line)}, true
	case strings.HasPrefix(line, ">======="):
		return LinesDiff{DiffMode: DiffConflictSplit, Line: asConflictLine(line)}, true
	case strings.HasPrefix(line, ">>>>>>>>"):
		return LinesDiff{DiffMode: DiffConflictEnd, Line: asConflictLine(line)}, true
	case strings.HasPrefix(line, ">"):
	case strings.HasPrefix(line, "<"):
	case strings.HasPrefix(line, "="):
func asConflictLine(line string) string {
	return strings.ReplaceAll(line[2:], "\t", "   ")
}