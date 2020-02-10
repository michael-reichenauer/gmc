package repoview

import (
	"fmt"
	"github.com/michael-reichenauer/gmc/repoview/viewmodel"
	"github.com/michael-reichenauer/gmc/utils"
	"github.com/michael-reichenauer/gmc/utils/gitlib"
	"github.com/michael-reichenauer/gmc/utils/log"
	"github.com/michael-reichenauer/gmc/utils/ui"
	"strings"
)

type diffVM struct {
	model        *viewmodel.Service
	currentIndex int
	filesDiff    []gitlib.FileDiff
	page         int
}

type diffPage struct {
	lines      []string
	firstIndex int
	total      int
}

func NewDiffVM(model *viewmodel.Service) *diffVM {
	log.Infof("new vm")
	return &diffVM{model: model, currentIndex: -1, page: 0}
}

func (h *diffVM) getCommitDiff(viewPort ui.ViewPage) (diffPage, error) {
	if h.currentIndex == -1 {
		log.Infof("Not set")
		return diffPage{}, nil
	}

	if h.filesDiff == nil {
		commit, err := h.model.GetCommitByIndex(h.currentIndex)
		if err != nil {
			log.Infof("commit not found")
			return diffPage{}, err
		}

		h.filesDiff, err = h.model.GetCommitDiff(commit.ID)
		if err != nil {
			return diffPage{}, err
		}
	}

	var lines []string
	lines = append(lines, utils.Text(fmt.Sprintf("Changed files: %d", len(h.filesDiff)), viewPort.Width))
	for _, df := range h.filesDiff {
		diffType := toDiffType(df)
		lines = append(lines, utils.Text(fmt.Sprintf("  %s %s", diffType, df.PathAfter), viewPort.Width))
	}
	lines = append(lines, utils.Text("", viewPort.Width))
	lines = append(lines, utils.Text("", viewPort.Width))
	for i, df := range h.filesDiff {
		if i != 0 {
			lines = append(lines, utils.Text("", viewPort.Width))
			lines = append(lines, utils.Text("", viewPort.Width))
		}
		lines = append(lines, ui.MagentaDk(strings.Repeat("═", viewPort.Width)))

		lines = append(lines, ui.Cyan(utils.Text(fmt.Sprintf("%s %s", toDiffType(df), df.PathAfter), viewPort.Width)))
		if df.IsRenamed {
			lines = append(lines, ui.Dark(utils.Text(fmt.Sprintf("Renamed: %s -> %s", df.PathBefore, df.PathAfter), viewPort.Width)))
		}
		for j, ds := range df.SectionDiffs {
			if j != 0 {
				//lines = append(lines, "")
				lines = append(lines, ui.Dark(strings.Repeat("─", viewPort.Width)))
			}
			linesText := fmt.Sprintf("Lines: %s", ds.ChangedIndexes)
			lines = append(lines, ui.Dark(utils.Text(linesText, viewPort.Width)))
			lines = append(lines, ui.Dark(strings.Repeat("─", viewPort.Width)))
			for _, dl := range ds.LinesDiffs {
				switch dl.DiffMode {
				case gitlib.DiffSame:
					lines = append(lines, utils.Text(fmt.Sprintf("  %s", dl.Line), viewPort.Width))
				case gitlib.DiffAdded:
					if h.page != -1 {
						lines = append(lines, ui.Green(utils.Text(fmt.Sprintf("+ %s", dl.Line), viewPort.Width)))
					}
				case gitlib.DiffRemoved:
					if h.page != 1 {
						lines = append(lines, ui.Red(utils.Text(fmt.Sprintf("- %s", dl.Line), viewPort.Width)))
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

func (h *diffVM) SetIndex(index int) {
	h.page = 0
	log.Infof("was %d", h.currentIndex)
	h.currentIndex = index
	h.filesDiff = nil
	log.Infof("is %d", h.currentIndex)
}

func (h *diffVM) SetLeft(page int) {
	h.page = page
}

func (h *diffVM) SetRight(page int) {
	h.page = page
}

func toDiffType(df gitlib.FileDiff) string {
	switch df.DiffMode {
	case gitlib.DiffModified:
		return "Modified:"
	case gitlib.DiffAdded:
		return "Added:   "
	case gitlib.DiffRemoved:
		return "Removed: "
	}
	return ""
}
