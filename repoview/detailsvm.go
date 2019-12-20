package repoview

import (
	"fmt"
	"github.com/michael-reichenauer/gmc/repoview/model"
	"github.com/michael-reichenauer/gmc/utils/ui"
	"strings"
)

type detailsVM struct {
	model         *model.Model
	selectedIndex int
}

type commitDetails struct {
	Text string
}

func newDetailsVM(model *model.Model) *detailsVM {
	return &detailsVM{model: model}
}

func (h detailsVM) getCommitDetails(viewPort ui.ViewPort, index int) (commitDetails, error) {
	commit, err := h.model.GetCommitByIndex(index)
	if err != nil {
		return commitDetails{}, err
	}
	return commitDetails{Text: toDetailsText(commit)}, nil
}

func toDetailsText(c model.Commit) string {

	var sb strings.Builder
	sb.WriteString(toHeader("Id:") + ui.Gray(c.ID))
	sb.WriteString("\n")
	sb.WriteString(toHeader("Branch:") + toBranchText(c))
	sb.WriteString("\n")
	if c.ID == model.StatusID {
		sb.WriteString(ui.YellowDk(" " + c.Message))
	} else {
		sb.WriteString(ui.Gray(" " + c.Message))
	}
	sb.WriteString("\n")
	return sb.String()
}

func toHeader(text string) string {
	return ui.White(fmt.Sprintf(" %-10s", text))
}
func toBranchText(c model.Commit) string {
	bColor := branchColor(c.Branch.DisplayName)
	typeText := ""
	//if c.Branch.IsRemote {
	//	typeText = ui.Dark("remote: ")
	//} else {
	//	typeText = ui.Dark("local: ")
	//}
	return typeText + ui.ColorText(bColor, c.Branch.DisplayName)
}
