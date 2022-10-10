package git

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"
	"unicode/utf8"

	"github.com/michael-reichenauer/gmc/utils"
	"github.com/michael-reichenauer/gmc/utils/log"
)

type DiffMode int

const (
	DiffModified DiffMode = iota
	DiffAdded
	DiffRemoved
	DiffSame
	DiffConflicts
	DiffConflictStart
	DiffConflictSplit
	DiffConflictEnd
)

type CommitDiff struct {
	Id        string
	Author    string
	Date      string
	Message   string
	FileDiffs []FileDiff
}

type FileDiff struct {
	PathBefore   string
	PathAfter    string
	IsRenamed    bool
	DiffMode     DiffMode
	SectionDiffs []SectionDiff
}

type SectionDiff struct {
	ChangedIndexes string
	LinesDiffs     []LinesDiff
}

type LinesDiff struct {
	DiffMode DiffMode
	Line     string
}

// fetches from remote origin
type diffService struct {
	cmd           gitCommander
	statusHandler *statusService
}

func newDiff(cmd gitCommander, statusHandler *statusService) *diffService {
	return &diffService{cmd: cmd, statusHandler: statusHandler}
}

func (t *diffService) commitDiff(id string) (CommitDiff, error) {
	if id == UncommittedID {
		return t.unCommittedDiff()
	}

	diffText, err := t.cmd.Git("show",
		"--first-parent", "--root", "--patch", "--ignore-space-change", "--no-color",
		//"--output-indicator-context==", "--output-indicator-new=>", "--output-indicator-old=<",
		"--find-renames", "--unified=6", id)
	if err != nil {
		return CommitDiff{}, err
	}
	commitDiffs, err := t.parse(diffText, false)
	if err != nil {
		return CommitDiff{}, err
	}
	return commitDiffs[0], nil
}

func (t *diffService) fileDiff(path string) ([]CommitDiff, error) {
	diffText, err := t.cmd.Git("log", "--patch", "--", path)
	if err != nil {
		return []CommitDiff{}, err
	}
	commitDiffs, err := t.parse(diffText, false)
	if err != nil {
		return []CommitDiff{}, err
	}
	return commitDiffs, nil
}

func (t *diffService) unCommittedDiff() (CommitDiff, error) {
	diffText, err := t.cmd.Git("diff",
		"--first-parent", "--root", "--patch", "--ignore-space-change", "--no-color",
		//	"--output-indicator-context==", "--output-indicator-new=>", "--output-indicator-old=<",
		"--find-renames", "--unified=6", "HEAD")
	if err != nil {
		return CommitDiff{}, err
	}

	commitDiffs, err := t.parse(diffText, true)
	if err != nil {
		return CommitDiff{}, err
	}

	fileDiffs := commitDiffs[0].FileDiffs
	status, err := t.statusHandler.getStatus()
	if err != nil {
		return CommitDiff{}, err
	}
	t.setConflictsFilesMode(fileDiffs, status)
	fileDiffs, err = t.addAddedFiles(fileDiffs, status, t.cmd.WorkingDir())
	if err != nil {
		return CommitDiff{}, err
	}
	sort.SliceStable(fileDiffs, func(i, j int) bool {
		return strings.Compare(strings.ToLower(fileDiffs[i].PathAfter), strings.ToLower(fileDiffs[j].PathAfter)) == -1
	})
	return CommitDiff{FileDiffs: fileDiffs}, err
}

func (t *diffService) parse(text string, isUncommitted bool) ([]CommitDiff, error) {
	lines := strings.Split(text, "\n")

	var currentCommitDiff *CommitDiff
	var currentFileDiff *FileDiff
	var currentSectionDiff *SectionDiff
	var commitDiffs []CommitDiff

	if isUncommitted {
		// Uncommitted diffs do not have a commit header
		commitDiffs = append(commitDiffs, CommitDiff{Id: UncommittedID, FileDiffs: []FileDiff{}})
		currentCommitDiff = &commitDiffs[len(commitDiffs)-1]
	}

	for i, line := range lines {
		if len(line) == 0 {
			continue
		}
		if cd, ok := tryParseCommitDiffHead(i, lines); ok {
			commitDiffs = append(commitDiffs, cd)
			currentCommitDiff = &commitDiffs[len(commitDiffs)-1]
			currentFileDiff = nil
			continue
		}
		if fd, ok := tryParseFileDiffHead(line); ok {
			currentCommitDiff.FileDiffs = append(currentCommitDiff.FileDiffs, fd)
			currentFileDiff = &currentCommitDiff.FileDiffs[len(currentCommitDiff.FileDiffs)-1]
			currentSectionDiff = nil
			continue
		}
		if fileDiffMode, ok := tryParseFileMode(line); ok && currentFileDiff != nil && currentFileDiff.DiffMode != DiffConflicts {
			currentFileDiff.DiffMode = fileDiffMode
			continue
		}
		if sd, ok := tryParseSectionHead(line); ok && currentFileDiff != nil {
			currentFileDiff.SectionDiffs = append(currentFileDiff.SectionDiffs, sd)
			currentSectionDiff = &currentFileDiff.SectionDiffs[len(currentFileDiff.SectionDiffs)-1]
			continue
		}
		if lineDiff, ok := tryParseLineDiff(line); ok && currentSectionDiff != nil {
			currentSectionDiff.LinesDiffs = append(currentSectionDiff.LinesDiffs, lineDiff)
			continue
		}
	}

	return commitDiffs, nil
}

func (t *diffService) addAddedFiles(diffs []FileDiff, status Status, dirPath string) ([]FileDiff, error) {
	for _, name := range status.AddedFiles {
		filePath := filepath.Join(dirPath, name)
		file, err := utils.FileRead(filePath)
		fileText := ""
		if err != nil {
			log.Warnf("Failed to read file %s, %t", filePath, err)
			fileText = fmt.Sprintf("<Error: %v>", err)
		} else if isText(file) {
			fileText = string(file)
		} else {
			fileText = fmt.Sprintf("<Not a text file: '%s'>", filePath)
		}

		lines := strings.Split(fileText, "\n")
		var lds []LinesDiff
		for _, line := range lines {
			line = strings.TrimRight(line, "\r")
			line = strings.ReplaceAll(line, "\t", "   ")
			lds = append(lds, LinesDiff{DiffMode: DiffAdded, Line: line})
		}
		sd := SectionDiff{ChangedIndexes: fmt.Sprintf("-0,0 +1,%d", len(lines)), LinesDiffs: lds}
		diffs = append(diffs, FileDiff{
			PathBefore:   name,
			PathAfter:    name,
			IsRenamed:    false,
			DiffMode:     DiffAdded,
			SectionDiffs: []SectionDiff{sd},
		})
	}
	return diffs, nil
}

func (t *diffService) setConflictsFilesMode(diffs []FileDiff, status Status) {
	for i, fd := range diffs {
		if utils.StringsContains(status.ConflictsFiles, fd.PathAfter) {
			fd.DiffMode = DiffConflicts
			diffs[i] = fd
		}
	}
}

func tryParseSectionHead(line string) (SectionDiff, bool) {
	var sectionDiff SectionDiff
	if !strings.HasPrefix(line, "@@") {
		return sectionDiff, false
	}
	endIndex := strings.Index(line[2:], "@@")
	if endIndex == -1 {
		return sectionDiff, false
	}
	sectionDiff.ChangedIndexes = line[3 : endIndex+1]
	return sectionDiff, true
}

func tryParseFileMode(line string) (DiffMode, bool) {
	if strings.HasPrefix(line, "new file mode") {
		return DiffAdded, true
	}
	if strings.HasPrefix(line, "deleted file mode") {
		return DiffRemoved, true
	}
	return DiffModified, false
}

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

func tryParseFileDiffHead(line string) (FileDiff, bool) {
	var fileDiff FileDiff
	if strings.HasPrefix(line, "diff --cc ") {
		return tryParseFileDiffConflictHead(line)
	}
	if !strings.HasPrefix(line, "diff --git ") {
		return fileDiff, false
	}
	fileDiff.DiffMode = DiffModified
	files := line[11:]
	parts := strings.Split(files, " ")
	fileDiff.PathBefore = parts[0][2:]
	fileDiff.PathAfter = parts[1][2:]
	fileDiff.IsRenamed = fileDiff.PathBefore != fileDiff.PathAfter
	return fileDiff, true
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

func tryParseLineDiff(line string) (LinesDiff, bool) {
	switch {
	case strings.HasPrefix(line, "+<<<<<<<"):
		return LinesDiff{DiffMode: DiffConflictStart, Line: asConflictLine(line)}, true
	case strings.HasPrefix(line, "+======="):
		return LinesDiff{DiffMode: DiffConflictSplit, Line: asConflictLine(line)}, true
	case strings.HasPrefix(line, "+>>>>>>>"):
		return LinesDiff{DiffMode: DiffConflictEnd, Line: asConflictLine(line)}, true
	case strings.HasPrefix(line, "+"):
		return LinesDiff{DiffMode: DiffAdded, Line: asLine(line)}, true
	case strings.HasPrefix(line, "-"):
		return LinesDiff{DiffMode: DiffRemoved, Line: asLine(line)}, true
	case strings.HasPrefix(line, " "):
		return LinesDiff{DiffMode: DiffSame, Line: asLine(line)}, true
	}
	return LinesDiff{}, false
}

func asLine(line string) string {
	return strings.ReplaceAll(line[1:], "\t", "   ")
}
func asConflictLine(line string) string {
	return strings.ReplaceAll(line[2:], "\t", "   ")
}

func isText(s []byte) bool {
	const max = 1024 // at least utf8.UTFMax
	if len(s) > max {
		s = s[0:max]
	}
	for i, c := range string(s) {
		if i+utf8.UTFMax > len(s) {
			// last char may be incomplete - ignore
			break
		}
		if c == 0xFFFD || c < ' ' && c != '\n' && c != '\t' && c != '\f' {
			// decoding error or control character - not a text file
			return false
		}
	}
	return true
}
