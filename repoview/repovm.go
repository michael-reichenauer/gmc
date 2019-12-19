package repoview

import (
	"fmt"
	"github.com/michael-reichenauer/gmc/repoview/model"
	"github.com/michael-reichenauer/gmc/utils/log"
	"github.com/michael-reichenauer/gmc/utils/ui"
	"strings"
)

const (
	RFC3339Small = "2006-01-02 15:04"
)

type repoPage struct {
	repoPath           string
	text               string
	lines              int
	currentBranchName  string
	currentCommitIndex int
	commitStatus       string
	first              int
	last               int
	current            int
}

type repoVM struct {
	currentCommit string
	repoPath      string
	model         *model.Model
	viewPort      model.ViewPort
	statusMessage string
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
	var err error
	h.viewPort, err = h.model.GetRepoViewPort(firstLine, lastLine, selected)
	if err != nil {
		return repoPage{}, err
	}
	firstLine = h.viewPort.First
	lastLine = h.viewPort.Last
	selected = h.viewPort.Selected

	markerWidth := 8 //13
	log.Infof("before width")
	messageLength, authorLength, timeLength := columnWidths(h.viewPort.GraphWidth+markerWidth, width)
	log.Infof("after width %d %d %d", messageLength, authorLength, timeLength)
	var sb strings.Builder
	commits := h.viewPort.Commits

	h.statusMessage = ""
	sbid := ""
	sbc := commits[selected-firstLine]
	sbid = sbc.Branch.Name

	for i, c := range commits {
		writeSelectedMarker(&sb, i+firstLine, selected)
		writeGraph(&sb, c)
		sb.WriteString(" ")
		writeMergeMarker(&sb, c)
		writeCurrentMarker(&sb, c)
		sb.WriteString(" ")
		writeMessage(&sb, c, sbid, messageLength)
		sb.WriteString(" ")
		writeAuthor(&sb, c, authorLength)
		sb.WriteString(" ")
		writeAuthorTime(&sb, c, timeLength)
		sb.WriteString("\n")
	}

	commitStatus := h.toCommitStatus(h.viewPort.Commits, selected-firstLine, h.viewPort.StatusMessage)

	return repoPage{
		repoPath:           h.repoPath,
		text:               sb.String(),
		lines:              h.viewPort.TotalCommits,
		currentBranchName:  h.viewPort.CurrentBranchName,
		currentCommitIndex: h.viewPort.CurrentCommitIndex,
		commitStatus:       commitStatus,
		first:              firstLine,
		last:               lastLine,
		current:            selected,
	}, nil
}

func (h *repoVM) toCommitStatus(commits []model.Commit, selected int, status string) string {
	if selected >= len(commits) {
		return ""
	}
	//if h.statusMessage != "" && selected == 0 {
	//	return ""
	//}
	//if h.statusMessage != "" {
	//	selected--
	//}
	c := commits[selected]
	return fmt.Sprintf(": %s   (%s %s)", status, c.SID, c.Branch.Name)
}

func (h *repoVM) OpenBranch(index int) {
	//if h.statusMessage != "" && index == 0 {
	//	return
	//}
	//if h.statusMessage != "" {
	//	index--
	//}
	h.model.OpenBranch(h.viewPort, index)
}

func (h *repoVM) CloseBranch(index int) {
	//if h.statusMessage != "" && index == 0 {
	//	return
	//}
	//if h.statusMessage != "" {
	//	index--
	//}
	h.model.CloseBranch(h.viewPort, index)
}

func (h *repoVM) Refresh() {
	//h.model.Refresh(h.viewPort)
}

func writeSelectedMarker(sb *strings.Builder, index, selected int) {
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
		bColor := branchColor(c.Graph[i].BranchDisplayName)

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
	sb.WriteString(ui.Dark(txt(commit.Author, length)))
}

func writeAuthorTime(sb *strings.Builder, commit model.Commit, length int) {
	sb.WriteString(ui.Dark(txt(commit.AuthorTime.Format(RFC3339Small), length)))
}

func writeMessage(sb *strings.Builder, c model.Commit, selectedBranchID string, length int) {
	messaged := txt(c.Message, length)
	if c.Branch.Name == selectedBranchID {
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
