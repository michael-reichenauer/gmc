package console

import (
	"fmt"
	"github.com/michael-reichenauer/gmc/api"
	"github.com/michael-reichenauer/gmc/utils/cui"
	"strings"
)

var (
	lineChar      = "─"
	branchPointer = "▲"
)

type detailsVM struct {
	currentCommit api.Commit
}

func NewDetailsVM() *detailsVM {
	return &detailsVM{}
}

func (h detailsVM) getCommitDetails(viewPort cui.ViewPage) (string, error) {
	return "", nil
	// width := viewPort.Width
	// c := h.currentCommit
	// var sb strings.Builder
	// sb.WriteString(h.toViewLine(width, c.Branch) + "\n")
	// id := c.ID
	// if c.IsUncommitted {
	// 	id = " "
	// }
	// sb.WriteString(toHeader("Id:") + cui.Dark(id) + "\n")
	// sb.WriteString(toHeader("Branch:") + h.toBranchText(c) + "\n")
	// sb.WriteString(toHeader("Files:") + cui.Dark("... >") + "\n")
	// sb.WriteString(toHeader("Parents:") + cui.Dark(toSids(c.ParentIDs)) + "\n")
	// sb.WriteString(toHeader("Children:") + cui.Dark(toSids(c.ChildIDs)) + "\n")
	// sb.WriteString(toHeader("Branch tips:") + cui.Dark(toBranchTips(c.BranchTips)) + "\n")
	//
	// color := cui.CDark
	// if c.IsUncommitted {
	// 	color = cui.CYellowDk
	// }
	//
	// messageLines := strings.Split(strings.TrimSpace(c.Message), "\n")
	// for i, line := range messageLines {
	// 	if i == 0 && len(messageLines) > 1 {
	// 		line = line + " >"
	// 	}
	// 	if i == 0 {
	// 		sb.WriteString(toHeader("Message:") + cui.ColorText(color, line) + "\n")
	// 	} else {
	// 		sb.WriteString("           " + cui.ColorText(color, line) + "\n")
	// 	}
	// }
	// return sb.String(), nil
}

func (h detailsVM) toViewLine(width int, branch api.Branch) string {
	prefixWidth := branch.Index*2 - 1
	suffixWidth := width - branch.Index*2 - 2
	pointer := " " + branchPointer + " "
	if prefixWidth < 0 {
		prefixWidth = 0
		pointer = branchPointer + " "
		suffixWidth++
	}
	return cui.Dark(strings.Repeat(lineChar, prefixWidth) +
		pointer +
		cui.Dark(strings.Repeat(lineChar, suffixWidth)))
}

func toBranchTips(tips []string) string {
	return fmt.Sprintf("%s", strings.Join(tips, ", "))
}

func toSids(ids []string) string {
	var sids []string
	for _, id := range ids {
		sids = append(sids, toSid(id))
	}
	return fmt.Sprintf("%s", strings.Join(sids, ", "))
}

func toSid(commitID string) string {
	return commitID[:6]
}

func toHeader(text string) string {
	return cui.White(fmt.Sprintf(" %-13s", text))
}

// func (h detailsVM) toBranchText(c api.Commit) string {
// 	typeText := ""
//
// 	switch {
// 	case c.Branch.IsMultiBranch:
// 		typeText = cui.Dark(" (multiple) >")
// 	case !c.Branch.IsGitBranch:
// 		typeText = cui.Dark(" ()")
// 	case c.IsUncommitted:
// 		typeText = cui.Dark(" (local)")
// 	case c.IsLocalOnly:
// 		typeText = cui.Dark(" (local)")
// 	case c.IsRemoteOnly:
// 		typeText = cui.Dark(" (remote)")
// 	case c.Branch.IsRemote && c.Branch.LocalName != "":
// 		typeText = cui.Dark(" (remote,local)")
// 	case !c.Branch.IsRemote && c.Branch.RemoteName != "":
// 		typeText = cui.Dark(" (remote,local)")
// 	case c.Branch.IsRemote && c.Branch.LocalName == "":
// 		typeText = cui.Dark(" (remote)")
// 	default:
// 		typeText = cui.Dark(" (local)")
// 	}
// 	if c.IsUncommitted {
// 		typeText = typeText + ", changes not yet committed"
// 	} else if c.IsRemoteOnly {
// 		typeText = typeText + ", commit not yet pulled"
// 	} else if c.IsLocalOnly {
// 		typeText = typeText + ", commit not yet pushed"
// 	}
// 	return c.Branch.DisplayName + cui.Dark(typeText)
// }

func (h detailsVM) setCurrentCommit(commit api.Commit) {
	h.currentCommit = commit
}
