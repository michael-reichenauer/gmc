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

var (
	lineChar      = "─"
	branchPointer = "▲"
)

type detailsVM struct {
	model         *viewmodel.Service
	selectedIndex int
}

type commitDetails struct {
	lines []string
}

type commitDiff struct {
	lines []string
}

func NewDetailsVM(model *viewmodel.Service) *detailsVM {
	return &detailsVM{model: model}
}

func (h detailsVM) getCommitDetails(viewPort ui.ViewPage, index int) (commitDetails, error) {
	commit, err := h.model.GetCommitByIndex(index)
	if err != nil {
		return commitDetails{}, err
	}
	return commitDetails{lines: h.toDetailsText(commit, viewPort.Width)}, nil
}

func (h detailsVM) getCommitDiff(viewPort ui.ViewPage, index int) (commitDiff, error) {
	log.Infof("geet diff")
	commit, err := h.model.GetCommitByIndex(index)
	if err != nil {
		log.Infof("commit not found")
		return commitDiff{}, err
	}
	diff, err := h.model.GetCommitDiff(commit.ID)

	var lines []string
	lines = append(lines, utils.Text(fmt.Sprintf("Changed files: %d", len(diff)), viewPort.Width))
	for _, df := range diff {
		diffType := toDiffType(df)
		lines = append(lines, utils.Text(fmt.Sprintf("  %s %s", diffType, df.PathAfter), viewPort.Width))
	}
	lines = append(lines, "")
	lines = append(lines, "")
	for i, df := range diff {
		if i != 0 {
			lines = append(lines, "")
			lines = append(lines, "")
		}
		lines = append(lines, ui.MagentaDk(strings.Repeat("═", viewPort.Width)))

		lines = append(lines, utils.Text(fmt.Sprintf("File:  %s", df.PathAfter), viewPort.Width))
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
					lines = append(lines, utils.Text(fmt.Sprintf("+ %s", dl.Line), viewPort.Width))
				case gitlib.DiffRemoved:
					lines = append(lines, utils.Text(fmt.Sprintf("- %s", dl.Line), viewPort.Width))
				}
			}
		}
	}
	return commitDiff{lines: lines}, nil
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

func (h detailsVM) toDetailsText(c viewmodel.Commit, width int) []string {
	var lines []string
	lines = append(lines, h.toViewLine(width, c.Branch))
	width = width - 14
	id := c.ID
	if id == viewmodel.StatusID {
		id = " "
	}
	lines = append(lines, toHeader("Id:")+ui.Dark(utils.Text(id, width)))
	lines = append(lines, toHeader("Branch:")+h.toBranchText(c, width))
	lines = append(lines, toHeader("Files:")+ui.Dark(utils.Text("... >", width)))
	lines = append(lines, toHeader("Parents:")+ui.Dark(utils.Text(toSids(c.ParentIDs), width)))
	lines = append(lines, toHeader("Children:")+ui.Dark(utils.Text(toSids(c.ChildIDs), width)))
	lines = append(lines, toHeader("Branch tips:")+ui.Dark(utils.Text(toBranchTips(c.BranchTips), width)))

	color := ui.CDark
	if c.ID == viewmodel.StatusID {
		color = ui.CYellowDk
	}

	messageLines := strings.Split(strings.TrimSpace(c.Message), "\n")
	for i, line := range messageLines {
		if i == 0 && len(messageLines) > 1 {
			line = line + " >"
		}
		line = utils.Text(line, width)
		if i == 0 {
			lines = append(lines, toHeader("Message:")+ui.ColorText(color, line))
		} else {
			lines = append(lines, "           "+ui.ColorText(color, line))
		}
	}
	return lines
}

func (h detailsVM) toViewLine(width int, branch viewmodel.Branch) string {
	bColor := h.model.BranchColor(branch.DisplayName)
	prefixWidth := branch.Index*2 - 1
	suffixWidth := width - branch.Index*2 - 2
	pointer := " " + branchPointer + " "
	if prefixWidth < 0 {
		prefixWidth = 0
		pointer = branchPointer + " "
		suffixWidth++
	}
	return ui.Dark(strings.Repeat(lineChar, prefixWidth) +
		ui.ColorText(bColor, pointer) +
		ui.Dark(strings.Repeat(lineChar, suffixWidth)))
}

func toBranchTips(tips []string) string {
	return fmt.Sprintf("%s", strings.Join(tips, ", "))
}

func toSids(ids []string) string {
	var sids []string
	for _, id := range ids {
		sids = append(sids, viewmodel.ToSid(id))
	}
	return fmt.Sprintf("%s", strings.Join(sids, ", "))
}

func toHeader(text string) string {
	return ui.White(fmt.Sprintf(" %-13s", text))
}

func (h detailsVM) toBranchText(c viewmodel.Commit, width int) string {
	bColor := h.model.BranchColor(c.Branch.DisplayName)
	typeText := ""

	switch {
	case c.Branch.IsMultiBranch:
		typeText = ui.Dark(" (multiple) >")
	case !c.Branch.IsGitBranch:
		typeText = ui.Dark(" ()")
	case c.ID == viewmodel.StatusID:
		typeText = ui.Dark(" (local)")
	case c.IsLocalOnly:
		typeText = ui.Dark(" (local)")
	case c.IsRemoteOnly:
		typeText = ui.Dark(" (remote)")
	case c.Branch.IsRemote && c.Branch.LocalName != "":
		typeText = ui.Dark(" (remote,local)")
	case !c.Branch.IsRemote && c.Branch.RemoteName != "":
		typeText = ui.Dark(" (remote,local)")
	case c.Branch.IsRemote && c.Branch.LocalName == "":
		typeText = ui.Dark(" (remote)")
	default:
		typeText = ui.Dark(" (local)")
	}
	if c.ID == viewmodel.StatusID {
		typeText = typeText + ", changes not yet committed"
	} else if c.IsRemoteOnly {
		typeText = typeText + ", commit not yet pulled"
	} else if c.IsLocalOnly {
		typeText = typeText + ", commit not yet pushed"
	}
	return ui.ColorText(bColor, c.Branch.DisplayName) +
		ui.Dark(utils.Text(typeText, width-len(c.Branch.DisplayName)+11))
}
