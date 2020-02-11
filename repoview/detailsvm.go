package repoview

import (
	"fmt"
	"github.com/michael-reichenauer/gmc/repoview/viewmodel"
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

func (h detailsVM) toDetailsText(c viewmodel.Commit, width int) []string {
	var lines []string
	lines = append(lines, h.toViewLine(width, c.Branch))
	id := c.ID
	if id == viewmodel.StatusID {
		id = " "
	}
	lines = append(lines, toHeader("Id:")+ui.Dark(id))
	lines = append(lines, toHeader("Branch:")+h.toBranchText(c))
	lines = append(lines, toHeader("Files:")+ui.Dark("... >"))
	lines = append(lines, toHeader("Parents:")+ui.Dark(toSids(c.ParentIDs)))
	lines = append(lines, toHeader("Children:")+ui.Dark(toSids(c.ChildIDs)))
	lines = append(lines, toHeader("Branch tips:")+ui.Dark(toBranchTips(c.BranchTips)))

	color := ui.CDark
	if c.ID == viewmodel.StatusID {
		color = ui.CYellowDk
	}

	messageLines := strings.Split(strings.TrimSpace(c.Message), "\n")
	for i, line := range messageLines {
		if i == 0 && len(messageLines) > 1 {
			line = line + " >"
		}
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

func (h detailsVM) toBranchText(c viewmodel.Commit) string {
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
	return ui.ColorText(bColor, c.Branch.DisplayName) + ui.Dark(typeText)
}
