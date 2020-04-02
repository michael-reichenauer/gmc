package repoview

import (
	"fmt"
	"github.com/michael-reichenauer/gmc/utils/git"
	"github.com/michael-reichenauer/gmc/utils/ui"
	"strings"
)

type DiffGetter interface {
	GetCommitDiff(id string) (git.CommitDiff, error)
}

type diffVM struct {
	diffViewer     ui.Viewer
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

func newDiffVM(diffViewer ui.Viewer, diffGetter DiffGetter, commitID string) *diffVM {
	return &diffVM{diffViewer: diffViewer, diffGetter: diffGetter, commitID: commitID}
}

func (h *diffVM) load() {
	go func() {
		diff, _ := h.diffGetter.GetCommitDiff(h.commitID)
		h.diffViewer.PostOnUIThread(func() {
			h.commitDiff = diff
			h.isDiffReady = true
			h.diffViewer.NotifyChanged()
		})
	}()
}

func (h *diffVM) setUnified(isUnified bool) {
	h.isUnified = isUnified
	h.isDiff = false
	h.leftLines = nil
	h.rightLines = nil
}

func (h *diffVM) getCommitDiffLeft(viewPort ui.ViewPage) (ui.ViewPageData, error) {
	return h.getCommitDiff(viewPort, true)
}

func (h *diffVM) getCommitDiffRight(viewPort ui.ViewPage) (ui.ViewPageData, error) {
	return h.getCommitDiff(viewPort, false)
}

func (h *diffVM) getCommitDiff(viewPort ui.ViewPage, isLeft bool) (ui.ViewPageData, error) {
	if !h.isDiffReady {
		return h.loadingText(isLeft), nil
	}
	if h.isNewDiffNeeded(viewPort.FirstCharIndex) {
		h.setDiffSides(viewPort.FirstCharIndex)
		h.isDiff = true
	}

	lines, firstIndex, lastIndex := h.getLines(isLeft, viewPort.FirstLine, viewPort.Height)

	return ui.ViewPageData{
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

func (h *diffVM) loadingText(isLeft bool) ui.ViewPageData {
	text := "Loading diff for " + h.commitID[:6]
	if !isLeft {
		text = ""
	}
	return ui.ViewPageData{Lines: []string{text}}
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
			h.addLeftAndRight(ui.Dark(strings.Repeat("─", viewWidth)))
		}
	}
	h.addLeftAndRight("")
}

func (h *diffVM) addDiffSummery() {
	h.addLeft(fmt.Sprintf("Changed files: %d", len(h.commitDiff.FileDiffs)))
	for _, df := range h.commitDiff.FileDiffs {
		diffType := h.toDiffType(df)
		h.addLeft(fmt.Sprintf("  %s %s", diffType, df.PathAfter))
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
	h.addLeftAndRight(ui.Blue(strings.Repeat("═", viewWidth)))
	fileText := ui.Cyan(fmt.Sprintf("%s %s", h.toDiffType(df), df.PathAfter))
	h.addLeftAndRight(fileText)
	if df.IsRenamed {
		renamedText := ui.Dark(fmt.Sprintf("Renamed: %s -> %s", df.PathBefore, df.PathAfter))
		h.addLeftAndRight(renamedText)
	}
}

func (h *diffVM) addDiffSectionHeader(ds git.SectionDiff) {
	h.addLeftAndRight("")
	leftLines, rightLines := h.parseLinesTexts(ds)
	h.add(ui.Dark(leftLines), ui.Dark(rightLines))
	h.addLeftAndRight(ui.Dark(strings.Repeat("─", viewWidth)))
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
	for _, dl := range ds.LinesDiffs {
		if len(dl.Line) > h.maxWidth {
			h.maxWidth = len(dl.Line)
		}
		l := h.line(dl.Line)
		switch dl.DiffMode {
		case git.DiffRemoved:
			leftBlock = append(leftBlock, ui.Red(l))
		case git.DiffAdded:
			rightBlock = append(rightBlock, ui.Green(l))
		case git.DiffSame:
			h.addBlocks(leftBlock, rightBlock)
			leftBlock = nil
			rightBlock = nil
			h.addLeftAndRight(l)
		}
	}
	h.addBlocks(leftBlock, rightBlock)
}

func (h *diffVM) addBlocks(left, right []string) {
	if h.isUnified {
		h.leftLines = append(h.leftLines, left...)
		h.leftLines = append(h.leftLines, right...)
		for i := 0; i < len(left)+len(right); i++ {
			h.rightLines = append(h.rightLines, ui.Dark(strings.Repeat("░", viewWidth)))
		}
		return
	}

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
		return "Modified:"
	case git.DiffAdded:
		return "Added:   "
	case git.DiffRemoved:
		return "Removed: "
	}
	return ""
}
