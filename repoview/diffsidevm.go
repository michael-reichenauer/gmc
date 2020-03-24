package repoview

import (
	"fmt"
	"github.com/michael-reichenauer/gmc/utils/git"
	"github.com/michael-reichenauer/gmc/utils/log"
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
	if !h.isDiffReady {
		log.Infof("Not set")
		return diffPage{lines: []string{"Loading diff for " + h.commitID}, firstIndex: 0, total: 1}, nil
	}
	if !h.isDiff {
		h.leftLines, h.rightLines = h.getCommitSides()
		h.isDiff = true
	}
	return h.getCommitDiff(viewPort, h.leftLines)
}

func (h *diffSideVM) getCommitDiffRight(viewPort ui.ViewPage) (diffPage, error) {
	if !h.isDiffReady {
		log.Infof("Not set")
		return diffPage{lines: []string{"Loading diff for " + h.commitID}, firstIndex: 0, total: 1}, nil
	}
	if !h.isDiff {
		h.leftLines, h.rightLines = h.getCommitSides()
		h.isDiff = true
	}
	return h.getCommitDiff(viewPort, h.rightLines)
}

func (h *diffSideVM) getCommitDiff(viewPort ui.ViewPage, lines []string) (diffPage, error) {
	if viewPort.FirstLine+viewPort.Height > len(lines) {
		viewPort.FirstLine = len(lines) - viewPort.Height
	}
	if viewPort.FirstLine < 0 {
		viewPort.FirstLine = 0
	}
	if viewPort.FirstLine+viewPort.Height > len(lines) {
		viewPort.Height = len(lines) - viewPort.FirstLine
	}

	return diffPage{
		lines:      lines[viewPort.FirstLine : viewPort.FirstLine+viewPort.Height],
		firstIndex: viewPort.FirstLine,
		total:      len(lines),
	}, nil
}

func (h *diffSideVM) getCommitSides() ([]string, []string) {
	viewWidth := 200
	var leftLines []string
	var rightLines []string
	leftLines, rightLines = h.add(leftLines, rightLines, fmt.Sprintf("Changed files: %d", len(h.commitDiff.FileDiffs)), "")

	for _, df := range h.commitDiff.FileDiffs {
		diffType := toDiffType(df)
		leftLines, rightLines = h.add(leftLines, rightLines, fmt.Sprintf("  %s %s", diffType, df.PathAfter), "")
	}

	leftLines, rightLines = h.add(leftLines, rightLines, "", "")
	leftLines, rightLines = h.add(leftLines, rightLines, "", "")

	for i, df := range h.commitDiff.FileDiffs {
		if i != 0 {
			leftLines, rightLines = h.add(leftLines, rightLines, "", "")
			leftLines, rightLines = h.add(leftLines, rightLines, "", "")
		}

		leftLines, rightLines = h.add(leftLines, rightLines, ui.Blue(strings.Repeat("═", viewWidth)), ui.Blue(strings.Repeat("═", viewWidth)))

		fileText := ui.Cyan(fmt.Sprintf("%s %s", toDiffType(df), df.PathAfter))
		leftLines, rightLines = h.add(leftLines, rightLines, fileText, fileText)
		if df.IsRenamed {
			renamedText := ui.Dark(fmt.Sprintf("Renamed: %s -> %s", df.PathBefore, df.PathAfter))
			leftLines, rightLines = h.add(leftLines, rightLines, renamedText, renamedText)
		}
		for _, ds := range df.SectionDiffs {
			leftLines, rightLines = h.add(leftLines, rightLines, ui.Dark(strings.Repeat("─", viewWidth)), ui.Dark(strings.Repeat("─", viewWidth)))
			linesText := fmt.Sprintf("Lines: %s", ds.ChangedIndexes)
			leftLines, rightLines = h.add(leftLines, rightLines, ui.Dark(linesText), ui.Dark(linesText))
			leftLines, rightLines = h.add(leftLines, rightLines, ui.Dark(strings.Repeat("─", viewWidth)), ui.Dark(strings.Repeat("─", viewWidth)))

			var lb []string
			var rb []string
			for _, dl := range ds.LinesDiffs {
				switch dl.DiffMode {
				case git.DiffRemoved:
					lb = append(lb, ui.Red(dl.Line))
					//leftLines, rightLines = h.add(leftLines, rightLines, ui.Red(dl.Line), "")
				case git.DiffAdded:
					rb = append(rb, ui.Green(dl.Line))
				//	leftLines, rightLines = h.add(leftLines, rightLines, "", ui.Green(dl.Line))
				case git.DiffSame:
					leftLines, rightLines = h.flushBlock(leftLines, rightLines, lb, rb)
					lb = nil
					rb = nil
					leftLines, rightLines = h.add(leftLines, rightLines, dl.Line, dl.Line)
				}
			}
			leftLines, rightLines = h.flushBlock(leftLines, rightLines, lb, rb)
			lb = nil
			rb = nil
		}
	}

	return leftLines, rightLines
}

func (h *diffSideVM) flushBlock(leftLines, rightLines []string, left, right []string) ([]string, []string) {
	leftLines = append(leftLines, left...)
	rightLines = append(rightLines, right...)
	if len(left) > len(right) {
		for i := 0; i < len(left)-len(right); i++ {
			rightLines = append(rightLines, "")
		}
	}
	if len(right) > len(left) {
		for i := 0; i < len(right)-len(left); i++ {
			leftLines = append(leftLines, "")
		}
	}
	return leftLines, rightLines
}

func (h *diffSideVM) add(leftLines, rightLines []string, left, right string) ([]string, []string) {
	leftLines = append(leftLines, left)
	rightLines = append(rightLines, right)
	return leftLines, rightLines
}
