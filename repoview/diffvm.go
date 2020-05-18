package repoview

import (
	"fmt"
	"github.com/michael-reichenauer/gmc/utils/cui"
	"github.com/michael-reichenauer/gmc/utils/git"
	"strings"
)

type DiffGetter interface {
	GetCommitDiff(id string) (git.CommitDiff, error)
}

type diffVM struct {
	viewer         cui.Viewer
	diffGetter     DiffGetter
	commitDiff     git.CommitDiff
	commitID       string
	isDiffReady    bool
	isDiff         bool
	leftLines      []string
	rightLines     []string
	isUnified      bool
	firstCharIndex int
	maxWidth       int
}

const viewWidth = 200

func newDiffVM(viewer cui.Viewer, diffGetter DiffGetter, commitID string) *diffVM {
	return &diffVM{viewer: viewer, diffGetter: diffGetter, commitID: commitID}
}

func (h *diffVM) load() {
	go func() {
		diff, _ := h.diffGetter.GetCommitDiff(h.commitID)
		h.viewer.PostOnUIThread(func() {
			h.commitDiff = diff
			h.isDiffReady = true
			h.viewer.NotifyChanged()
		})
	}()
}

func (h *diffVM) setUnified(isUnified bool) {
	h.isUnified = isUnified
	h.isDiff = false
	h.leftLines = nil
	h.rightLines = nil
}

func (h *diffVM) getCommitDiffLeft(viewPort cui.ViewPage) (cui.ViewText, error) {
	return h.getCommitDiff(viewPort, true)
}

func (h *diffVM) getCommitDiffRight(viewPort cui.ViewPage) (cui.ViewText, error) {
	return h.getCommitDiff(viewPort, false)
}

func (h *diffVM) getCommitDiff(viewPort cui.ViewPage, isLeft bool) (cui.ViewText, error) {
	if !h.isDiffReady {
		return h.loadingText(isLeft), nil
	}
	if h.isNewDiffNeeded(viewPort.FirstCharIndex) {
		h.setDiffSides(viewPort.FirstCharIndex)
		h.isDiff = true
	}

	lines, firstIndex, lastIndex := h.getLines(isLeft, viewPort.FirstLine, viewPort.Height)

	return cui.ViewText{
		Lines:    lines[firstIndex:lastIndex],
		Total:    len(lines),
		MaxWidth: h.maxWidth,
	}, nil
}

func (h *diffVM) isNewDiffNeeded(firstCharIndex int) bool {
	return !h.isDiff || h.firstCharIndex != firstCharIndex
}

func (h *diffVM) getLines(isLeft bool, firstIndex, height int) ([]string, int, int) {
	lines := h.leftLines
	lastIndex := firstIndex + height
	if !isLeft {
		lines = h.rightLines
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

func (h *diffVM) loadingText(isLeft bool) cui.ViewText {
	text := "Loading diff for " + h.commitID[:6]
	if !isLeft {
		text = ""
	}
	return cui.ViewText{Lines: []string{text}}
}

func (h *diffVM) setDiffSides(firstCharIndex int) {
	h.leftLines = nil
	h.rightLines = nil
	h.maxWidth = 0
	// Adding diff summery with changed files list, count, ...
	h.firstCharIndex = firstCharIndex
	h.addDiffSummery()

	// Add file diffs
	for _, df := range h.commitDiff.FileDiffs {
		h.addFileHeader(df)

		// Add all diff sections in a file
		for _, ds := range df.SectionDiffs {
			h.addDiffSectionHeader(ds)
			h.addDiffSectionLines(ds)
			h.addLeftAndRight(cui.Dark(strings.Repeat("─", viewWidth)))
		}
	}
	h.addLeftAndRight("")
}

func (h *diffVM) addDiffSummery() {
	h.addLeft(fmt.Sprintf("%d Files:", len(h.commitDiff.FileDiffs)))
	for _, df := range h.commitDiff.FileDiffs {
		diffType := h.toDiffType(df)
		if df.DiffMode == git.DiffConflicts {
			h.addLeft(cui.Yellow(fmt.Sprintf("  %s %s", diffType, df.PathAfter)))
		} else {
			h.addLeft(fmt.Sprintf("  %s %s", diffType, df.PathAfter))
		}
	}
}

func (h *diffVM) line(text string) string {
	if h.firstCharIndex > len(text) {
		return ""
	}
	if h.firstCharIndex <= 0 {
		return text
	}
	return text[h.firstCharIndex:]
}

func (h *diffVM) addFileHeader(df git.FileDiff) {
	h.addLeftAndRight("")
	h.addLeftAndRight("")
	h.addLeftAndRight(cui.Blue(strings.Repeat("═", viewWidth)))
	fileText := cui.Cyan(fmt.Sprintf("%s %s", h.toDiffType(df), df.PathAfter))
	h.addLeftAndRight(fileText)
	if df.IsRenamed {
		renamedText := cui.Dark(fmt.Sprintf("Renamed: %s -> %s", df.PathBefore, df.PathAfter))
		h.addLeftAndRight(renamedText)
	}
}

func (h *diffVM) addDiffSectionHeader(ds git.SectionDiff) {
	h.addLeftAndRight("")
	leftLines, rightLines := h.parseLinesTexts(ds)
	h.add(cui.Dark(leftLines), cui.Dark(rightLines))
	h.addLeftAndRight(cui.Dark(strings.Repeat("─", viewWidth)))
}

func (h *diffVM) parseLinesTexts(ds git.SectionDiff) (string, string) {
	if h.isUnified {
		return fmt.Sprintf("Lines %s:", ds.ChangedIndexes), ""
	}

	parts := strings.Split(ds.ChangedIndexes, "+")
	leftText := fmt.Sprintf("Lines %s:", strings.TrimSpace(parts[0][1:]))
	rightText := fmt.Sprintf("Lines %s:", strings.TrimSpace(parts[1]))
	return leftText, rightText
}

func (h *diffVM) addDiffSectionLines(ds git.SectionDiff) {
	var leftBlock []string
	var rightBlock []string
	diffMode := git.DiffConflictEnd
	for _, dl := range ds.LinesDiffs {
		if len(dl.Line) > h.maxWidth {
			h.maxWidth = len(dl.Line)
		}
		l := h.line(dl.Line)

		switch dl.DiffMode {
		case git.DiffConflictStart:
			diffMode = git.DiffConflictStart
			h.addBlocks(leftBlock, rightBlock)
			leftBlock = nil
			rightBlock = nil
			h.addLeftAndRight(cui.Dark("=== Start of conflict "))
		case git.DiffConflictSplit:
			diffMode = git.DiffConflictSplit
		case git.DiffConflictEnd:
			diffMode = git.DiffConflictEnd
			h.addBlocks(leftBlock, rightBlock)
			leftBlock = nil
			rightBlock = nil
			h.addLeftAndRight(cui.Dark("=== End of conflict "))
		case git.DiffRemoved:
			if diffMode == git.DiffConflictStart {
				leftBlock = append(leftBlock, cui.Yellow(l))
			} else if diffMode == git.DiffConflictSplit {
				rightBlock = append(rightBlock, cui.Yellow(l))
			} else {
				leftBlock = append(leftBlock, cui.Red(l))
			}
		case git.DiffAdded:
			if diffMode == git.DiffConflictStart {
				leftBlock = append(leftBlock, cui.Yellow(l))
			} else if diffMode == git.DiffConflictSplit {
				rightBlock = append(rightBlock, cui.Yellow(l))
			} else {
				rightBlock = append(rightBlock, cui.Green(l))
			}
		case git.DiffSame:
			if diffMode == git.DiffConflictStart {
				leftBlock = append(leftBlock, cui.Yellow(l))
			} else if diffMode == git.DiffConflictSplit {
				rightBlock = append(rightBlock, cui.Yellow(l))
			} else {
				h.addBlocks(leftBlock, rightBlock)
				leftBlock = nil
				rightBlock = nil
				h.addLeftAndRight(l)
			}
		}
	}
	h.addBlocks(leftBlock, rightBlock)
}

func (h *diffVM) addBlocks(left, right []string) {
	if h.isUnified {
		h.leftLines = append(h.leftLines, left...)
		h.leftLines = append(h.leftLines, right...)
		for i := 0; i < len(left)+len(right); i++ {
			h.rightLines = append(h.rightLines, cui.Dark(strings.Repeat("░", viewWidth)))
		}
		return
	}

	h.leftLines = append(h.leftLines, left...)
	h.rightLines = append(h.rightLines, right...)
	if len(left) > len(right) {
		for i := 0; i < len(left)-len(right); i++ {
			h.rightLines = append(h.rightLines, cui.Dark(strings.Repeat("░", viewWidth)))
		}
	}
	if len(right) > len(left) {
		for i := 0; i < len(right)-len(left); i++ {
			h.leftLines = append(h.leftLines, cui.Dark(strings.Repeat("░", viewWidth)))
		}
	}
}

func (h *diffVM) addLeftAndRight(text string) {
	h.add(text, text)
}

func (h *diffVM) addLeft(left string) {
	h.add(left, "")
}

func (h *diffVM) add(left, right string) {
	h.leftLines = append(h.leftLines, left)
	h.rightLines = append(h.rightLines, right)
}

func (h *diffVM) toDiffType(df git.FileDiff) string {
	switch df.DiffMode {
	case git.DiffModified:
		return "Modified:  "
	case git.DiffAdded:
		return "Added:     "
	case git.DiffRemoved:
		return "Removed:   "
	case git.DiffConflicts:
		return "Conflicted:"
	}
	return ""
}
