package console

import (
	"fmt"
	"strings"

	"github.com/michael-reichenauer/gmc/api"
	"github.com/michael-reichenauer/gmc/utils/cui"
	"github.com/michael-reichenauer/gmc/utils/git"
)

type DiffGetter interface {
	GetCommitDiff(info api.CommitDiffInfoReq, diff *api.CommitDiff) error
}

type diffVM struct {
	ui             cui.UI
	viewer         cui.Viewer
	diffGetter     DiffGetter
	commitDiff     api.CommitDiff
	commitID       string
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

func newDiffVM(ui cui.UI, viewer cui.Viewer, diffGetter DiffGetter, repoID string, commitID string) *diffVM {
	return &diffVM{ui: ui, viewer: viewer, diffGetter: diffGetter, repoID: repoID, commitID: commitID}
}

func (t *diffVM) load() {
	progress := t.ui.ShowProgress("Getting diff ...")

	go func() {
		var diff api.CommitDiff
		err := t.diffGetter.GetCommitDiff(api.CommitDiffInfoReq{RepoID: t.repoID, CommitID: t.commitID}, &diff)
		t.viewer.PostOnUIThread(func() {
			progress.Close()
			if err != nil {
				t.ui.ShowErrorMessageBox("Failed to get diff:\n%v", err)
				return
			}
			t.commitDiff = diff
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
	text := "Loading diff for " + t.commitID[:6]
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
	t.addDiffSummery()

	// Add file diffs
	for _, df := range t.commitDiff.FileDiffs {
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

func (t *diffVM) addDiffSummery() {
	t.addLeft(fmt.Sprintf("%d Files:", len(t.commitDiff.FileDiffs)))
	for _, df := range t.commitDiff.FileDiffs {
		diffType := t.toDiffType(df)
		if df.DiffMode == api.DiffConflicts {
			t.addLeft(cui.Yellow(fmt.Sprintf("  %s %s", diffType, df.PathAfter)))
		} else {
			t.addLeft(fmt.Sprintf("  %s %s", diffType, df.PathAfter))
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

func (t *diffVM) addFileHeader(df api.FileDiff) {
	t.addLeftAndRight("")
	t.addLeftAndRight("")
	t.addLeftAndRight(cui.Blue(strings.Repeat("═", viewWidth)))
	fileText := cui.Cyan(fmt.Sprintf("%s %s", t.toDiffType(df), df.PathAfter))
	t.addLeftAndRight(fileText)
	if df.IsRenamed {
		renamedText := cui.Dark(fmt.Sprintf("Renamed: %s -> %s", df.PathBefore, df.PathAfter))
		t.addLeftAndRight(renamedText)
	}
}

func (t *diffVM) addDiffSectionHeader(df api.FileDiff, ds api.SectionDiff) {
	t.addLeftAndRight("")
	leftLines, rightLines := t.parseLinesTexts(df, ds)
	t.add(cui.Dark(leftLines), cui.Dark(rightLines))
	t.addLeftAndRight(cui.Dark(strings.Repeat("─", viewWidth)))
}

func (t *diffVM) parseLinesTexts(df api.FileDiff, ds api.SectionDiff) (string, string) {
	if t.isUnified {
		return fmt.Sprintf("%s:%s:", df.PathAfter, ds.ChangedIndexes), ""
	}

	parts := strings.Split(ds.ChangedIndexes, "+")
	leftLines := strings.Replace(strings.TrimSpace(parts[0][1:]), ",", " to ", 1)
	rightLines := strings.Replace(strings.TrimSpace(parts[1]), ",", " to ", 1)
	leftText := fmt.Sprintf("%s:%s:", df.PathBefore, leftLines)
	rightText := fmt.Sprintf("%s:%s:", df.PathAfter, rightLines)
	return leftText, rightText
}

func (t *diffVM) addDiffSectionLines(ds api.SectionDiff) {
	var leftBlock []string
	var rightBlock []string
	diffMode := git.DiffConflictEnd
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
				leftBlock = append(leftBlock, cui.Red(l))
			}
		case api.DiffAdded:
			if diffMode == git.DiffConflictStart {
				leftBlock = append(leftBlock, cui.Yellow(l))
			} else if diffMode == git.DiffConflictSplit {
				rightBlock = append(rightBlock, cui.Yellow(l))
			} else {
				rightBlock = append(rightBlock, cui.Green(l))
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
				t.addLeftAndRight(l)
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
