package repoview

import (
	"gmc/repoview/model"
	"gmc/utils/ui"
	"strings"
)

const (
	RFC3339Small = "2006-01-02 15:04"
)

var branchColors = []ui.Color{
	ui.CRed,
	ui.CBlue,
	ui.CYellow,
	ui.CGreen,
	ui.CCyan,
}
var (
	currentMarker         = ui.White("●")
	currentMarkerSelected = ui.White("┥")
	selectedMarker2       = ui.White("│")
	selectedMarker        = '│'
	moreMarker            = ui.Dark(">")
)

type repoPage struct {
	repoPath           string
	text               string
	lines              int
	currentBranchName  string
	currentCommitIndex int
}

type repoVM struct {
	currentCommit string
	repoPath      string
	model         *model.Model
}

func newRepoVM(repoPath string) *repoVM {
	return &repoVM{
		currentCommit: "",
		repoPath:      repoPath,
		model:         model.NewModel(repoPath),
	}
}

func (h *repoVM) Load() {
	h.model.Load()
}

func (h *repoVM) GetRepoPage(width, firstLine, lastLine, selected int) (repoPage, error) {
	vp, err := h.model.GetRepoViewPort(firstLine, lastLine, selected)
	if err != nil {
		return repoPage{}, err
	}

	if selected > lastLine {
		selected = lastLine
	}

	markerWidth := 13
	messageLength, authorLength, timeLength := columnWidths(vp.GraphWidth+markerWidth, width)

	var sb strings.Builder
	for i, c := range vp.Commits {
		writeSelectedMarker(&sb, c, i+firstLine, selected)
		writeGraph(&sb, c)
		sb.WriteString(" ")
		writeMergeMarker(&sb, c)
		//writeCurrentMarker2(&sb, c, i+firstLine, selected)
		writeCurrentMarker(&sb, c)
		sb.WriteString(" ")
		writeMessage(&sb, c, vp.SelectedBranch.ID, messageLength)
		sb.WriteString(" ")
		writeSid(&sb, c)
		sb.WriteString(" ")
		writeAuthor(&sb, c, authorLength)
		sb.WriteString(" ")
		writeAuthorTime(&sb, c, timeLength)
		sb.WriteString("\n")
	}

	return repoPage{
		repoPath:           h.repoPath,
		text:               sb.String(),
		lines:              vp.TotalCommits,
		currentBranchName:  vp.CurrentBranchName,
		currentCommitIndex: vp.CurrentCommitIndex,
	}, nil
}

func (h *repoVM) OpenBranch(index int) {
	h.model.OpenBranch(index)
}

func (h *repoVM) CloseBranch(index int) {
	h.model.CloseBranch(index)
}

func writeSelectedMarker(sb *strings.Builder, c model.Commit, index, selected int) {
	if index == selected {
		//color := branchColor(c.Branch.ID)
		color := ui.CWhite
		sb.WriteString(ui.ColorText(color, selectedMarker))
	} else {
		sb.WriteString(" ")
	}
}
func writeMergeMarker(sb *strings.Builder, c model.Commit) {
	if c.IsMore {
		sb.WriteString(moreMarker)
	} else {
		sb.WriteString(" ")
	}
}
func writeGraph(sb *strings.Builder, c model.Commit) {
	for i := 0; i < len(c.Graph); i++ {
		bColor := branchColor(c.Graph[i].BranchId)

		if i != 0 {
			cColor := bColor
			if c.Graph[i].Connect.Has(model.BPass) {
				cColor = ui.CWhite
			}
			sb.WriteString(ui.ColorText(cColor, graphConnectRune(c.Graph[i].Connect)))
		}
		if c.Graph[i].Branch == model.BPass {
			bColor = ui.CWhite
		}
		sb.WriteString(ui.ColorText(bColor, graphBranchRune(c.Graph[i].Branch)))
	}
}
func writeCurrentMarker(sb *strings.Builder, c model.Commit) {
	if c.IsCurrent {
		sb.WriteString(currentMarker)
	} else {
		sb.WriteString(" ")
	}
}
func writeCurrentMarker2(sb *strings.Builder, c model.Commit, index, selected int) {
	if c.IsCurrent && index == selected {
		sb.WriteString(currentMarkerSelected)
	} else if index == selected {
		sb.WriteString(selectedMarker2)
	} else if c.IsCurrent {
		sb.WriteString(currentMarker)
	} else {
		sb.WriteString(" ")
	}
}
func columnWidths(graphWidth, viewWidth int) (msgLength int, authorLength int, timeLength int) {
	width := viewWidth - graphWidth
	authorLength = 20
	timeLength = 16
	if width < 80 {
		authorLength = 10
		timeLength = 10
	}
	if width < 50 {
		authorLength = 0
		timeLength = 0
	}
	msgLength = viewWidth - graphWidth - authorLength - timeLength
	return
}
func writeSid(sb *strings.Builder, c model.Commit) {
	sb.WriteString(ui.Dark(c.SID))
}

func writeAuthor(sb *strings.Builder, commit model.Commit, length int) {
	sb.WriteString(ui.Dark(txt(commit.Author, length)))
}

func writeAuthorTime(sb *strings.Builder, commit model.Commit, length int) {
	sb.WriteString(ui.Dark(txt(commit.AuthorTime.Format(RFC3339Small), length)))
}

func writeMessage(sb *strings.Builder, c model.Commit, selectedBranchID string, length int) {
	messaged := txt(c.Message, length)
	if c.Branch.ID == selectedBranchID {
		sb.WriteString(ui.White(messaged))
	} else {
		sb.WriteString(ui.Dark(messaged))
	}
}

func txt(text string, length int) string {
	if len(text) <= length {
		return text + strings.Repeat(" ", length-len(text))
	}
	return text[0:length]
}
