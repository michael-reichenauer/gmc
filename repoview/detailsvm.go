package repoview

import (
	"fmt"
	"github.com/michael-reichenauer/gmc/repoview/viewmodel"
	"github.com/michael-reichenauer/gmc/utils"
	"github.com/michael-reichenauer/gmc/utils/ui"
	"strings"
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
	width = width - 14
	var lines []string
	lines = append(lines, toHeader("Id:")+ui.Dark(utils.Text(c.ID, width)))
	lines = append(lines, toHeader("Branch:")+toBranchText(c, width))
	lines = append(lines, toHeader("Files:")+ui.Dark(utils.Text("... >", width)))

	color := ui.CDark
	if c.ID == viewmodel.StatusID {
		color = ui.CYellowDk
	}

	for i, line := range strings.Split(c.Message, "\n") {
		line = utils.Text(line, width)
		if i == 0 {
			lines = append(lines, toHeader("Message:")+ui.ColorText(color, line))
		} else {
			lines = append(lines, "           "+ui.ColorText(color, line))
		}
	}

	return lines
}

func toHeader(text string) string {
	return ui.White(fmt.Sprintf(" %-10s", text))
}

func toBranchText(c viewmodel.Commit, width int) string {
	bColor := branchColor(c.Branch.DisplayName)
	typeText := ""
	//if c.Branch.IsRemote {
	//	typeText = ui.Dark("remote: ")
	//} else {
	//	typeText = ui.Dark("local: ")
	//}
	return typeText + ui.ColorText(bColor, utils.Text(c.Branch.DisplayName, width))
}
