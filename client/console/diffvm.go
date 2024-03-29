package console

import (
	"fmt"
	"strings"

	"github.com/michael-reichenauer/gmc/api"
	"github.com/michael-reichenauer/gmc/utils/cui"
	"github.com/michael-reichenauer/gmc/utils/git"
)

type DiffGetter interface {
	GetCommitDiff(info api.CommitDiffInfoReq) (api.CommitDiff, error)
	GetFileDiff(info api.FileDiffInfoReq) ([]api.CommitDiff, error)
}

type diffVM struct {
	ui             cui.UI
	viewer         cui.Viewer
	diffGetter     DiffGetter
	commitDiffs    []api.CommitDiff
	commitID       string
	path           string
	isDiffReady    bool
	isDiff         bool
	leftLines      []string
	rightLines     []string
	isUnified      bool
	firstCharIndex int
	maxWidth       int
	repoID         string
}

const viewWidth = 200

func newCommitDiffVM(ui cui.UI, viewer cui.Viewer, diffGetter DiffGetter, repoID string, commitID string) *diffVM {
	return &diffVM{ui: ui, viewer: viewer, diffGetter: diffGetter, repoID: repoID, commitID: commitID}
}

func newFileDiffVM(ui cui.UI, viewer cui.Viewer, diffGetter DiffGetter, repoID string, path string) *diffVM {
	return &diffVM{ui: ui, viewer: viewer, diffGetter: diffGetter, repoID: repoID, path: path}
}

func (t *diffVM) load() {
	if t.commitID != "" {
		t.loadCommitDiff()
		return
	}

	t.loadFileDiff()
}

func (t *diffVM) loadCommitDiff() {
	progress := t.ui.ShowProgress("Getting diff ...")

	go func() {
		diff, err := t.diffGetter.GetCommitDiff(api.CommitDiffInfoReq{RepoID: t.repoID, CommitID: t.commitID})
		t.viewer.PostOnUIThread(func() {
			progress.Close()
			if err != nil {
				t.ui.ShowErrorMessageBox("Failed to get diff:\n%v", err)
				return
			}
			t.commitDiffs = []api.CommitDiff{diff}
			t.isDiffReady = true
			t.viewer.NotifyChanged()
		})
	}()
}

func (t *diffVM) loadFileDiff() {
	progress := t.ui.ShowProgress("Getting diff ...")

	go func() {
		diff, err := t.diffGetter.GetFileDiff(api.FileDiffInfoReq{RepoID: t.repoID, Path: t.path})
		t.viewer.PostOnUIThread(func() {
			progress.Close()
			if err != nil {
				t.ui.ShowErrorMessageBox("Failed to get diff:\n%v", err)
				return
			}
			t.commitDiffs = diff
			t.isDiffReady = true
			t.viewer.NotifyChanged()
		})
	}()
}

func (t *diffVM) setUnified(isUnified bool) {
	t.isUnified = isUnified
	t.isDiff = false
	t.leftLines = nil
	t.rightLines = nil
}

func (t *diffVM) getCommitDiffLeft(viewPort cui.ViewPage) (cui.ViewText, error) {
	return t.getCommitDiff(viewPort, true)
}

func (t *diffVM) getCommitDiffRight(viewPort cui.ViewPage) (cui.ViewText, error) {
	return t.getCommitDiff(viewPort, false)
}

func (t *diffVM) getCommitDiff(viewPort cui.ViewPage, isLeft bool) (cui.ViewText, error) {
	if !t.isDiffReady {
		return t.loadingText(isLeft), nil
	}
	if t.isNewDiffNeeded(viewPort.FirstCharIndex) {
		t.setDiffSides(viewPort.FirstCharIndex)
		t.isDiff = true
	}

	lines, firstIndex, lastIndex := t.getLines(isLeft, viewPort.FirstLine, viewPort.Height)

	return cui.ViewText{
		Lines:    lines[firstIndex:lastIndex],
		Total:    len(lines),
		MaxWidth: t.maxWidth,
	}, nil
}

func (t *diffVM) isNewDiffNeeded(firstCharIndex int) bool {
	return !t.isDiff || t.firstCharIndex != firstCharIndex
}

func (t *diffVM) getLines(isLeft bool, firstIndex, height int) ([]string, int, int) {
	lines := t.leftLines
	lastIndex := firstIndex + height
	if !isLeft {
		lines = t.rightLines
	}

	if firstIndex+height > len(lines) {
		firstIndex = len(lines) - height
	}
	if firstIndex < 0 {
		firstIndex = 0
	}
	if firstIndex+height > len(lines) {
		lastIndex = firstIndex + len(lines) - firstIndex
	}
	return lines, firstIndex, lastIndex
}

func (t *diffVM) loadingText(isLeft bool) cui.ViewText {
	text := "Loading diff"
	if !isLeft {
		text = ""
	}
	return cui.ViewText{Lines: []string{text}}
}

func (t *diffVM) setDiffSides(firstCharIndex int) {
	t.leftLines = nil
	t.rightLines = nil
	t.maxWidth = 0
	// Adding diff summery with changed files list, count, ...
	t.firstCharIndex = firstCharIndex

	for _, commitDiff := range t.commitDiffs {
		t.addDiffSummery(commitDiff)

		// Add file diffs
		for _, df := range commitDiff.FileDiffs {
			t.addFileHeader(df)

			// Add all diff sections in a file
			for _, ds := range df.SectionDiffs {
				t.addDiffSectionHeader(df, ds)
				t.addDiffSectionLines(ds)
				t.addLeftAndRight(cui.Dark(strings.Repeat("─", viewWidth)))
			}
		}
		t.addLeftAndRight("")
	}
}

func (t *diffVM) addDiffSummery(commitDiff api.CommitDiff) {
	t.addLeftAndRight(cui.YellowDk(strings.Repeat("═", viewWidth)))
	if commitDiff.Id == "" {
		t.addLeft("Commit: uncommitted")
	} else {
		t.addLeft(fmt.Sprintf("Commit:  %s", commitDiff.Id[:6]))
		t.addLeft(fmt.Sprintf("Author:  %s", commitDiff.Author))
		t.addLeft(fmt.Sprintf("Date:    %s", commitDiff.Date))
		t.addLeft(fmt.Sprintf("Message: %s", commitDiff.Message))
	}

	if t.commitID != "" {
		t.addLeft("")
		t.addLeft(fmt.Sprintf("%d Files:", len(commitDiff.FileDiffs)))
		for _, df := range commitDiff.FileDiffs {
			diffType := t.toDiffType(df)
			if df.DiffMode == api.DiffConflicts {
				t.addLeft(cui.Yellow(fmt.Sprintf("  %s %s", diffType, df.PathAfter)))
			} else if df.DiffMode == api.DiffAdded {
				t.addLeft(cui.Green(fmt.Sprintf("  %s %s", diffType, df.PathAfter)))
			} else if df.DiffMode == api.DiffRemoved {
				t.addLeft(cui.Red(fmt.Sprintf("  %s %s", diffType, df.PathAfter)))
			} else {
				if df.IsRenamed {
					t.addLeft(fmt.Sprintf("  %s %s", diffType, df.PathAfter) + cui.Dark(fmt.Sprintf(" (renamed from %s)", df.PathBefore)))
				} else {
					t.addLeft(fmt.Sprintf("  %s %s", diffType, df.PathAfter))
				}
			}
		}
	}
}

func (t *diffVM) line(text string) string {
	if t.firstCharIndex > len(text) {
		return ""
	}
	if t.firstCharIndex <= 0 {
		return text
	}
	return text[t.firstCharIndex:]
}

func (t *diffVM) lineWithNr(lineNr int, text string, color cui.Color) string {
	lineNrText := fmt.Sprintf("%4d ", lineNr)

	if t.firstCharIndex > len(text)+len(lineNrText) {
		return ""
	}
	if t.firstCharIndex <= 0 {
		// Return whole row with line nr and line text
		return cui.Dark(lineNrText) + cui.ColorText(color, text)
	}
	if t.firstCharIndex <= len(lineNrText) {
		// Return partial line nr and whole text line
		return cui.Dark(lineNrText[t.firstCharIndex:]) + cui.ColorText(color, text)
	}
	// Return no line nr and partial text line
	return cui.ColorText(color, text[t.firstCharIndex-len(lineNrText):])
}

func (t *diffVM) addFileHeader(df api.FileDiff) {
	t.addLeftAndRight("")
	t.addLeftAndRight("")
	t.addLeftAndRight(cui.Blue(strings.Repeat("─", viewWidth)))
	fileText := cui.Cyan(fmt.Sprintf("%s %s", t.toDiffType(df), df.PathAfter))
	t.addLeftAndRight(fileText)
	if df.IsRenamed {
		renamedText := cui.Dark(fmt.Sprintf("Renamed:    %s -> %s", df.PathBefore, df.PathAfter))
		t.addLeftAndRight(renamedText)
	}
}

func (t *diffVM) addDiffSectionHeader(df api.FileDiff, ds api.SectionDiff) {
	t.addLeftAndRight("")
	t.addLeftAndRight(cui.Dark(strings.Repeat("─", viewWidth)))
}

func (t *diffVM) addDiffSectionLines(ds api.SectionDiff) {
	var leftBlock []string
	var rightBlock []string
	diffMode := git.DiffConflictEnd
	leftNr := ds.LeftLine
	rightNr := ds.RightLine
	for _, dl := range ds.LinesDiffs {
		if len(dl.Line) > t.maxWidth {
			t.maxWidth = len(dl.Line)
		}
		l := t.line(dl.Line)

		switch dl.DiffMode {
		case api.DiffConflictStart:
			diffMode = git.DiffConflictStart
			t.addBlocks(leftBlock, rightBlock)
			leftBlock = nil
			rightBlock = nil
			t.addLeftAndRight(cui.Dark("=== Start of conflict "))
		case api.DiffConflictSplit:
			diffMode = git.DiffConflictSplit
		case api.DiffConflictEnd:
			diffMode = git.DiffConflictEnd
			t.addBlocks(leftBlock, rightBlock)
			leftBlock = nil
			rightBlock = nil
			t.addLeftAndRight(cui.Dark("=== End of conflict "))

		case api.DiffRemoved:
			if diffMode == git.DiffConflictStart {
				leftBlock = append(leftBlock, cui.Yellow(l))
			} else if diffMode == git.DiffConflictSplit {
				rightBlock = append(rightBlock, cui.Yellow(l))
			} else {
				lnr := t.lineWithNr(leftNr, dl.Line, cui.CRed)
				leftNr++
				leftBlock = append(leftBlock, lnr)
			}

		case api.DiffAdded:
			if diffMode == git.DiffConflictStart {
				leftBlock = append(leftBlock, cui.Yellow(l))
			} else if diffMode == git.DiffConflictSplit {
				rightBlock = append(rightBlock, cui.Yellow(l))
			} else {
				lnr := t.lineWithNr(rightNr, dl.Line, cui.CGreen)
				rightNr++
				rightBlock = append(rightBlock, lnr)
			}

		case api.DiffSame:
			if diffMode == git.DiffConflictStart {
				leftBlock = append(leftBlock, cui.Yellow(l))
			} else if diffMode == git.DiffConflictSplit {
				rightBlock = append(rightBlock, cui.Yellow(l))
			} else {
				t.addBlocks(leftBlock, rightBlock)
				leftBlock = nil
				rightBlock = nil

				lnr := t.lineWithNr(leftNr, dl.Line, cui.CWhite)
				leftNr++
				rnr := t.lineWithNr(rightNr, dl.Line, cui.CWhite)
				rightNr++
				t.add(lnr, rnr)
			}
		}
	}
	t.addBlocks(leftBlock, rightBlock)
}

func (t *diffVM) addBlocks(left, right []string) {
	if t.isUnified {
		t.leftLines = append(t.leftLines, left...)
		t.leftLines = append(t.leftLines, right...)
		for i := 0; i < len(left)+len(right); i++ {
			t.rightLines = append(t.rightLines, cui.Dark(strings.Repeat("░", viewWidth)))
		}
		return
	}

	t.leftLines = append(t.leftLines, left...)
	t.rightLines = append(t.rightLines, right...)
	if len(left) > len(right) {
		for i := 0; i < len(left)-len(right); i++ {
			t.rightLines = append(t.rightLines, cui.Dark(strings.Repeat("░", viewWidth)))
		}
	}
	if len(right) > len(left) {
		for i := 0; i < len(right)-len(left); i++ {
			t.leftLines = append(t.leftLines, cui.Dark(strings.Repeat("░", viewWidth)))
		}
	}
}

func (t *diffVM) addLeftAndRight(text string) {
	t.add(text, text)
}

func (t *diffVM) addLeft(left string) {
	t.add(left, "")
}

func (t *diffVM) add(left, right string) {
	t.leftLines = append(t.leftLines, left)
	t.rightLines = append(t.rightLines, right)
}

func (t *diffVM) toDiffType(df api.FileDiff) string {
	switch df.DiffMode {
	case api.DiffModified:
		return "Modified:  "
	case api.DiffAdded:
		return "Added:     "
	case api.DiffRemoved:
		return "Removed:   "
	case api.DiffConflicts:
		return "Conflicted:"
	}
	return ""
}
