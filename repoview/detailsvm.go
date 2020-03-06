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
	currentCommit viewmodel.Commit
}

func NewDetailsVM() *detailsVM {
	return &detailsVM{}
}

func (h detailsVM) getCommitDetails(viewPort ui.ViewPage) (string, error) {
	width := viewPort.Width
	c := h.currentCommit
	var sb strings.Builder
	sb.WriteString(h.toViewLine(width, c.Branch) + "\n")
	id := c.ID
	if id == viewmodel.StatusID {
		id = " "
	}
	sb.WriteString(toHeader("Id:") + ui.Dark(id) + "\n")
	sb.WriteString(toHeader("Branch:") + h.toBranchText(c) + "\n")
	sb.WriteString(toHeader("Files:") + ui.Dark("... >") + "\n")
	sb.WriteString(toHeader("Parents:") + ui.Dark(toSids(c.ParentIDs)) + "\n")
	sb.WriteString(toHeader("Children:") + ui.Dark(toSids(c.ChildIDs)) + "\n")
	sb.WriteString(toHeader("Branch tips:") + ui.Dark(toBranchTips(c.BranchTips)) + "\n")

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
			sb.WriteString(toHeader("Message:") + ui.ColorText(color, line) + "\n")
		} else {
			sb.WriteString("           " + ui.ColorText(color, line) + "\n")
		}
	}
	return sb.String(), nil
}

func (h detailsVM) toViewLine(width int, branch viewmodel.Branch) string {
	prefixWidth := branch.Index*2 - 1
	suffixWidth := width - branch.Index*2 - 2
	pointer := " " + branchPointer + " "
	if prefixWidth < 0 {
		prefixWidth = 0
		pointer = branchPointer + " "
		suffixWidth++
	}
	return ui.Dark(strings.Repeat(lineChar, prefixWidth) +
		pointer +
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
	return c.Branch.DisplayName + ui.Dark(typeText)
}

func (h detailsVM) setCurrentCommit(commit viewmodel.Commit) {
	h.currentCommit = commit
}
