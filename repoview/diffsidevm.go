package repoview

import (
	"fmt"
	"github.com/michael-reichenauer/gmc/utils/git"
	"github.com/michael-reichenauer/gmc/utils/ui"
	"strings"
)

type diffSideVM struct {
	diffViewer  ui.Viewer
	diffGetter  DiffGetter
	commitDiff  git.CommitDiff
	commitID    string
	isDiffReady bool
	isDiff      bool
	leftLines   []string
	rightLines  []string
}

const viewWidth = 200

func NewDiffSideVM(diffViewer ui.Viewer, diffGetter DiffGetter, commitID string) *diffSideVM {
	return &diffSideVM{diffViewer: diffViewer, diffGetter: diffGetter, commitID: commitID}
}

func (h *diffSideVM) load() {
	go func() {
		diff, _ := h.diffGetter.GetCommitDiff(h.commitID)
		h.diffViewer.PostOnUIThread(func() {
			h.commitDiff = diff
			h.isDiffReady = true
			h.diffViewer.NotifyChanged()
		})
	}()
}

func (h *diffSideVM) getCommitDiffLeft(viewPort ui.ViewPage) (diffPage, error) {
	return h.getCommitDiff(viewPort, true)
}

func (h *diffSideVM) getCommitDiffRight(viewPort ui.ViewPage) (diffPage, error) {
	return h.getCommitDiff(viewPort, false)
}

func (h *diffSideVM) getCommitDiff(viewPort ui.ViewPage, isLeft bool) (diffPage, error) {
	if !h.isDiffReady {
		return h.loadingText(isLeft), nil
	}
	if !h.isDiff {
		h.setCommitSides()
		h.isDiff = true
	}

	lines, firstIndex, lastIndex := h.getLines(isLeft, viewPort.FirstLine, viewPort.Height)

	return diffPage{
		lines:      lines[firstIndex:lastIndex],
		firstIndex: firstIndex,
		total:      len(lines),
	}, nil
}

func (h *diffSideVM) getLines(isLeft bool, firstIndex, height int) ([]string, int, int) {
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

func (h *diffSideVM) loadingText(isLeft bool) diffPage {
	text := "Loading diff for " + h.commitID[:6]
	if !isLeft {
		text = ""
	}
	return diffPage{lines: []string{text}, firstIndex: 0, total: 1}
}

func (h *diffSideVM) setCommitSides() {
	// Adding diff summery with changed files list, count, ...
	h.addDiffSummery()

	// Add file diffs
	for _, df := range h.commitDiff.FileDiffs {
		h.addFileHeader(df)

		// Add all diff sections in a file
		for _, ds := range df.SectionDiffs {
			h.addDiffSectionHeader(ds)
			h.addDiffSectionLines(ds)
		}
	}
}

func (h *diffSideVM) addDiffSummery() {
	h.addLeft(fmt.Sprintf("Changed files: %d", len(h.commitDiff.FileDiffs)))
	for _, df := range h.commitDiff.FileDiffs {
		diffType := toDiffType(df)
		h.addLeft(fmt.Sprintf("  %s %s", diffType, df.PathAfter))
	}
}

func (h *diffSideVM) addFileHeader(df git.FileDiff) {
	h.addLeftAndRight("")
	h.addLeftAndRight("")
	h.addLeftAndRight(ui.Blue(strings.Repeat("═", viewWidth)))
	fileText := ui.Cyan(fmt.Sprintf("%s %s", toDiffType(df), df.PathAfter))
	h.addLeftAndRight(fileText)
	if df.IsRenamed {
		renamedText := ui.Dark(fmt.Sprintf("Renamed: %s -> %s", df.PathBefore, df.PathAfter))
		h.addLeftAndRight(renamedText)
	}
}

func (h *diffSideVM) addDiffSectionHeader(ds git.SectionDiff) {
	h.addLeftAndRight(ui.Dark(strings.Repeat("─", viewWidth)))
	leftLines, rightLines := h.parseLinesTexts(ds)
	h.add(ui.Dark(leftLines), ui.Dark(rightLines))
	h.addLeftAndRight(ui.Dark(strings.Repeat("─", viewWidth)))
}

func (h *diffSideVM) parseLinesTexts(ds git.SectionDiff) (string, string) {
	parts := strings.Split(ds.ChangedIndexes, "+")
	leftText := fmt.Sprintf("Lines: %s", strings.TrimSpace(parts[0][1:]))
	rightText := fmt.Sprintf("Lines: %s", strings.TrimSpace(parts[1]))
	return leftText, rightText
}

func (h *diffSideVM) addDiffSectionLines(ds git.SectionDiff) {
	var leftBlock []string
	var rightBlock []string
	for _, dl := range ds.LinesDiffs {
		switch dl.DiffMode {
		case git.DiffRemoved:
			leftBlock = append(leftBlock, ui.Red(dl.Line))
		case git.DiffAdded:
			rightBlock = append(rightBlock, ui.Green(dl.Line))
		case git.DiffSame:
			h.addBlocks(leftBlock, rightBlock)
			leftBlock = nil
			rightBlock = nil
			h.addLeftAndRight(dl.Line)
		}
	}
	h.addBlocks(leftBlock, rightBlock)
}

func (h *diffSideVM) addBlocks(left, right []string) {
	h.leftLines = append(h.leftLines, left...)
	h.rightLines = append(h.rightLines, right...)
	if len(left) > len(right) {
		for i := 0; i < len(left)-len(right); i++ {
			h.rightLines = append(h.rightLines, ui.Dark(strings.Repeat("░", viewWidth)))
		}
	}
	if len(right) > len(left) {
		for i := 0; i < len(right)-len(left); i++ {
			h.leftLines = append(h.leftLines, ui.Dark(strings.Repeat("░", viewWidth)))
		}
	}
}

func (h *diffSideVM) addLeftAndRight(text string) {
	h.add(text, text)
}

func (h *diffSideVM) addLeft(left string) {
	h.add(left, "")
}

func (h *diffSideVM) addRight(right string) {
	h.add("", right)
}

func (h *diffSideVM) add(left, right string) {
	h.leftLines = append(h.leftLines, left)
	h.rightLines = append(h.rightLines, right)
}
