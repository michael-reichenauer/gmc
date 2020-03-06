package repoview

import (
	"fmt"
	"github.com/michael-reichenauer/gmc/utils"
	"github.com/michael-reichenauer/gmc/utils/git"
	"github.com/michael-reichenauer/gmc/utils/log"
	"github.com/michael-reichenauer/gmc/utils/ui"
	"strings"
)

type diffVM struct {
	diffViewer  ui.Viewer
	diffGetter  diffGetter
	commitDiff  git.CommitDiff
	commitID    string
	isDiffReady bool
	page        int
}

type diffPage struct {
	lines      []string
	firstIndex int
	total      int
	title      string
}

type diffGetter interface {
	GetCommitDiff(id string) git.CommitDiff
}

func NewDiffVM(diffViewer ui.Viewer, diffGetter diffGetter, commitID string) *diffVM {
	log.Infof("new vm")
	return &diffVM{diffViewer: diffViewer, diffGetter: diffGetter, commitID: commitID}
}

func (h *diffVM) load() {
	go func() {
		diff := h.diffGetter.GetCommitDiff(h.commitID)
		h.diffViewer.PostOnUIThread(func() {
			h.commitDiff = diff
			h.isDiffReady = true
			h.diffViewer.NotifyChanged()
		})
	}()
}

func (h *diffVM) onLeft() {
	if h.page >= 0 {
		h.page--
	} else {
		return
	}
	h.diffViewer.NotifyChanged()
}

func (h *diffVM) onRight() {
	if h.page <= 0 {
		h.page++
	} else {
		return
	}
	h.diffViewer.NotifyChanged()
}

func (h *diffVM) getCommitDiff(viewPort ui.ViewPage) (diffPage, error) {
	if !h.isDiffReady {
		log.Infof("Not set")
		return diffPage{lines: []string{"Loading diff for " + h.commitID}, firstIndex: 0, total: 1, title: "Diff"}, nil
	}

	var lines []string
	lines = append(lines, utils.Text(fmt.Sprintf("Changed files: %d", len(h.commitDiff.FileDiffs)), viewPort.Width))
	for _, df := range h.commitDiff.FileDiffs {
		diffType := toDiffType(df)
		lines = append(lines, utils.Text(fmt.Sprintf("  %s %s", diffType, df.PathAfter), viewPort.Width))
	}
	lines = append(lines, utils.Text("", viewPort.Width))
	lines = append(lines, utils.Text("", viewPort.Width))
	for i, df := range h.commitDiff.FileDiffs {
		if i != 0 {
			lines = append(lines, utils.Text("", viewPort.Width))
			lines = append(lines, utils.Text("", viewPort.Width))
		}
		lines = append(lines, ui.MagentaDk(strings.Repeat("═", viewPort.Width)))

		lines = append(lines, ui.Cyan(fmt.Sprintf("%s %s", toDiffType(df), df.PathAfter)))
		if df.IsRenamed {
			lines = append(lines, ui.Dark(fmt.Sprintf("Renamed: %s -> %s", df.PathBefore, df.PathAfter)))
		}
		for j, ds := range df.SectionDiffs {
			if j != 0 {
				//lines = append(lines, "")
				lines = append(lines, ui.Dark(strings.Repeat("─", viewPort.Width)))
			}
			linesText := fmt.Sprintf("Lines: %s", ds.ChangedIndexes)
			lines = append(lines, ui.Dark(linesText))
			lines = append(lines, ui.Dark(strings.Repeat("─", viewPort.Width)))
			for _, dl := range ds.LinesDiffs {
				switch dl.DiffMode {
				case git.DiffSame:
					lines = append(lines, fmt.Sprintf("  %s", dl.Line))
				case git.DiffAdded:
					if h.page != -1 {
						lines = append(lines, ui.Green(fmt.Sprintf("> %s", dl.Line)))
					}
				case git.DiffRemoved:
					if h.page != 1 {
						lines = append(lines, ui.Red(fmt.Sprintf("< %s", dl.Line)))
					}
				}
			}
		}
	}

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

func toDiffType(df git.FileDiff) string {
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
