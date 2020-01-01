package repoview

import (
	"fmt"
	"github.com/michael-reichenauer/gmc/repoview/viewmodel"
	"github.com/michael-reichenauer/gmc/utils"
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

func newDetailsVM(model *viewmodel.Service) *detailsVM {
	return &detailsVM{model: model}
}

func (h detailsVM) getCommitDetails(viewPort ui.ViewPage, index int) (commitDetails, error) {
	commit, err := h.model.GetCommitByIndex(index)
	if err != nil {
		return commitDetails{}, err
	}
	return commitDetails{lines: toDetailsText(commit, viewPort.Width)}, nil
}

func toDetailsText(c viewmodel.Commit, width int) []string {
	var lines []string
	lines = append(lines, toViewLine(width, c.Branch))
	width = width - 14
	id := c.ID
	if id == viewmodel.StatusID {
		id = " "
	}
	lines = append(lines, toHeader("Id:")+ui.Dark(utils.Text(id, width)))
	lines = append(lines, toHeader("Branch:")+toBranchText(c, width))
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

func toViewLine(width int, branch viewmodel.Branch) string {
	bColor := branchColor(branch.DisplayName)
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

func toBranchText(c viewmodel.Commit, width int) string {
	bColor := branchColor(c.Branch.DisplayName)
	typeText := ""

	switch {
	case c.Branch.IsMultiBranch:
		typeText = ui.Dark(" (multiple) >")
	case !c.Branch.IsGitBranch:
		typeText = ui.Dark(" (deleted)")
	case c.Branch.IsRemote && c.Branch.LocalName != "":
		typeText = ui.Dark(" (remote,local)")
	case c.Branch.IsRemote && c.Branch.LocalName == "":
		typeText = ui.Dark(" (remote)")
	default:
		typeText = ui.Dark(" (local)")
	}
	return ui.ColorText(bColor, c.Branch.DisplayName) +
		ui.Dark(utils.Text(typeText, width-len(c.Branch.DisplayName)+11))
}
