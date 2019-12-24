package repoview

import (
	"fmt"
	"github.com/michael-reichenauer/gmc/repoview/viewmodel"
	"github.com/michael-reichenauer/gmc/utils/ui"
)

type detailsVM struct {
	model         *viewmodel.Model
	selectedIndex int
}

type commitDetails struct {
	lines []string
}

func newDetailsVM(model *viewmodel.Model) *detailsVM {
	return &detailsVM{model: model}
}

func (h detailsVM) getCommitDetails(viewPort ui.ViewPort, index int) (commitDetails, error) {
	commit, err := h.model.GetCommitByIndex(index)
	if err != nil {
		return commitDetails{}, err
	}
	return commitDetails{lines: toDetailsText(commit)}, nil
}

func toDetailsText(c viewmodel.Commit) []string {

	var lines []string
	lines = append(lines, toHeader("Id:")+ui.Gray(c.ID))

	lines = append(lines, toHeader("Branch:")+toBranchText(c))

	if c.ID == viewmodel.StatusID {
		lines = append(lines, ui.YellowDk(" "+c.Message))
	} else {
		lines = append(lines, ui.Gray(" "+c.Message))
	}

	return lines
}

func toHeader(text string) string {
	return ui.White(fmt.Sprintf(" %-10s", text))
}
func toBranchText(c viewmodel.Commit) string {
	bColor := branchColor(c.Branch.DisplayName)
	typeText := ""
	//if c.Branch.IsRemote {
	//	typeText = ui.Dark("remote: ")
	//} else {
	//	typeText = ui.Dark("local: ")
	//}
	return typeText + ui.ColorText(bColor, c.Branch.DisplayName)
}
