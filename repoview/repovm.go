package repoview

import (
	"github.com/michael-reichenauer/gmc/repoview/model"
	"github.com/michael-reichenauer/gmc/utils"
	"github.com/michael-reichenauer/gmc/utils/ui"
	"strings"
)

const (
	RFC3339Small = "2006-01-02 15:04"
	markerWidth  = 8
)

type repoPage struct {
	repoPath           string
	text               string
	lines              int
	currentBranchName  string
	currentCommitIndex int
	first              int
	last               int
	current            int
	uncommittedChanges int
}

type repoVM struct {
	currentCommit string
	model         *model.Model
	viewPort      model.ViewPort
}

func newRepoVM(model *model.Model) *repoVM {
	return &repoVM{
		currentCommit: "",
		model:         model,
	}
}

func (h *repoVM) Load() {
	h.model.Load()
}

func (h *repoVM) GetRepoPage(width, firstLine, lastLine, selected int) (repoPage, error) {
	var err error
	h.viewPort, err = h.model.GetRepoViewPort(firstLine, lastLine, selected)
	if err != nil {
		return repoPage{}, err
	}
	firstLine = h.viewPort.First
	lastLine = h.viewPort.Last
	selected = h.viewPort.Selected

	messageLength, authorLength, timeLength := columnWidths(h.viewPort.GraphWidth+markerWidth, width)
	var sb strings.Builder
	commits := h.viewPort.Commits

	selectedCommit := commits[selected-firstLine]

	for i, c := range commits {
		writeSelectedMarker(&sb, i+firstLine, selected)
		writeGraph(&sb, c)
		sb.WriteString(" ")
		writeMergeMarker(&sb, c)
		writeCurrentMarker(&sb, c)
		sb.WriteString(" ")
		writeSubject(&sb, c, selectedCommit, messageLength)
		sb.WriteString(" ")
		writeAuthor(&sb, c, authorLength)
		sb.WriteString(" ")
		writeAuthorTime(&sb, c, timeLength)
		sb.WriteString("\n")
	}

	return repoPage{
		repoPath:           h.viewPort.RepoPath,
		text:               sb.String(),
		lines:              h.viewPort.TotalCommits,
		currentBranchName:  h.viewPort.CurrentBranchName,
		currentCommitIndex: h.viewPort.CurrentCommitIndex,
		first:              firstLine,
		last:               lastLine,
		current:            selected,
		uncommittedChanges: h.viewPort.UncommittedChanges,
	}, nil
}

func (h *repoVM) OpenBranch(index int) {
	h.model.OpenBranch(h.viewPort, index)
}

func (h *repoVM) CloseBranch(index int) {
	h.model.CloseBranch(h.viewPort, index)
}

func (h *repoVM) Refresh() {
	//h.model.Refresh(h.viewPort)
}

func writeSelectedMarker(sb *strings.Builder, index, selected int) {
	if index == selected {
		//color := branchColor(c.Branch.ID)
		color := ui.CWhite
		sb.WriteString(ui.ColorRune(color, selectedMarker))
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
		bColor := branchColor(c.Graph[i].BranchDisplayName)

		if i != 0 {
			cColor := bColor
			if c.Graph[i].Connect.Has(model.BPass) {
				cColor = ui.CWhite
			}
			sb.WriteString(ui.ColorRune(cColor, graphConnectRune(c.Graph[i].Connect)))
		}
		if c.Graph[i].Branch == model.BPass {
			bColor = ui.CWhite
		}
		sb.WriteString(ui.ColorRune(bColor, graphBranchRune(c.Graph[i].Branch)))
	}
}
func writeCurrentMarker(sb *strings.Builder, c model.Commit) {
	if c.IsCurrent {
		sb.WriteString(currentCommitMarker)
	} else {
		sb.WriteString(" ")
	}
}

func columnWidths(graphWidth, viewWidth int) (msgLength int, authorLength int, timeLength int) {
	width := viewWidth - graphWidth
	authorLength = 20
	timeLength = 16
	if width < 90 {
		authorLength = 10
		timeLength = 10
	}
	if width < 60 {
		authorLength = 0
		timeLength = 0
	}
	msgLength = viewWidth - graphWidth - authorLength - timeLength
	if msgLength < 0 {
		msgLength = 0
	}
	return
}
func writeSid(sb *strings.Builder, c model.Commit) {
	sb.WriteString(ui.Dark(c.SID))
}

func writeAuthor(sb *strings.Builder, commit model.Commit, length int) {
	sb.WriteString(ui.Dark(utils.Text(commit.Author, length)))
}

func writeAuthorTime(sb *strings.Builder, c model.Commit, length int) {
	if c.ID == model.StatusID {
		return
	}
	sb.WriteString(ui.Dark(utils.Text(c.AuthorTime.Format(RFC3339Small), length)))
}

func writeSubject(sb *strings.Builder, c model.Commit, selectedCommit model.Commit, length int) {
	subject := utils.Text(c.Subject, length)
	if c.ID == model.StatusID {
		sb.WriteString(ui.YellowDk(subject))
		return
	}
	if c.Branch.Name == selectedCommit.Branch.Name ||
		c.Branch.Name == selectedCommit.Branch.RemoteName ||
		c.Branch.RemoteName == selectedCommit.Branch.Name {
		sb.WriteString(ui.White(subject))
	} else {
		sb.WriteString(ui.Dark(subject))
	}
}
