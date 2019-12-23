package repoview

import (
	"github.com/michael-reichenauer/gmc/repoview/viewmodel"
	"github.com/michael-reichenauer/gmc/utils"
	"github.com/michael-reichenauer/gmc/utils/log"
	"github.com/michael-reichenauer/gmc/utils/ui"
	"strings"
)

const (
	RFC3339Small = "2006-01-02 15:04"
	markerWidth  = 8
)

type repoPage struct {
	lines              []string
	firstIndex         int
	currentIndex       int
	total              int
	repoPath           string
	currentBranchName  string
	uncommittedChanges int
}

type repoVM struct {
	notifier notifier
	model    *viewmodel.Model
	viewPort viewmodel.ViewPort
}

type notifier interface {
	NotifyChanged()
}

func newRepoVM(model *viewmodel.Model, notifier notifier) *repoVM {
	return &repoVM{
		notifier: notifier,
		model:    model,
	}
}

func (h *repoVM) Load() {
	log.Infof("repovm viewData ...")
	h.model.Start()
	h.model.TriggerRefresh()
	go h.monitorModelRoutine()
}

func (h *repoVM) monitorModelRoutine() {
	for range h.model.ChangedEvents {
		h.notifier.NotifyChanged()
	}
}

func (h *repoVM) GetRepoPage(viewPort ui.ViewPort) (repoPage, error) {
	var err error
	h.viewPort, err = h.model.GetRepoViewPort(viewPort.FirstIndex, viewPort.Height)
	if err != nil {
		return repoPage{}, err
	}
	messageLength, authorLength, timeLength := columnWidths(h.viewPort.GraphWidth+markerWidth, viewPort.Width)

	commits := h.viewPort.Commits

	var currentLineCommit viewmodel.Commit
	if viewPort.CurrentIndex-viewPort.FirstIndex < len(commits) && viewPort.CurrentIndex-viewPort.FirstIndex >= 0 {
		currentLineCommit = commits[viewPort.CurrentIndex-viewPort.FirstIndex]
	}

	var lines []string
	for _, c := range commits {
		var sb strings.Builder
		writeGraph(&sb, c)
		sb.WriteString(" ")
		writeMoreMarker(&sb, c)
		writeCurrentMarker(&sb, c)
		sb.WriteString(" ")
		writeSubject(&sb, c, currentLineCommit, messageLength)
		sb.WriteString(" ")
		writeAuthor(&sb, c, authorLength)
		sb.WriteString(" ")
		writeAuthorTime(&sb, c, timeLength)
		lines = append(lines, sb.String())
	}

	return repoPage{
		repoPath:           h.viewPort.RepoPath,
		lines:              lines,
		total:              h.viewPort.TotalCommits,
		firstIndex:         h.viewPort.FirstIndex,
		currentIndex:       viewPort.CurrentIndex,
		uncommittedChanges: h.viewPort.UncommittedChanges,
		currentBranchName:  h.viewPort.CurrentBranchName,
	}, nil
}

func (h *repoVM) OpenBranch(index int) {
	h.model.OpenBranch(index)
}

func (h *repoVM) CloseBranch(index int) {
	h.model.CloseBranch(index)
}

func (h *repoVM) Refresh() {
	//h.viewmodel.Refresh(h.viewPort)
}

func writeMoreMarker(sb *strings.Builder, c viewmodel.Commit) {
	if c.IsMore {
		sb.WriteString(moreMarker)
	} else {
		sb.WriteString(" ")
	}
}

func writeGraph(sb *strings.Builder, c viewmodel.Commit) {
	for i := 0; i < len(c.Graph); i++ {
		bColor := branchColor(c.Graph[i].BranchDisplayName)

		if i != 0 {
			cColor := bColor
			if c.Graph[i].Connect.Has(viewmodel.BPass) {
				cColor = ui.CWhite
			}
			sb.WriteString(ui.ColorRune(cColor, graphConnectRune(c.Graph[i].Connect)))
		}
		if c.Graph[i].Branch == viewmodel.BPass {
			bColor = ui.CWhite
		}
		sb.WriteString(ui.ColorRune(bColor, graphBranchRune(c.Graph[i].Branch)))
	}
}
func writeCurrentMarker(sb *strings.Builder, c viewmodel.Commit) {
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

func writeAuthor(sb *strings.Builder, commit viewmodel.Commit, length int) {
	sb.WriteString(ui.Dark(utils.Text(commit.Author, length)))
}

func writeAuthorTime(sb *strings.Builder, c viewmodel.Commit, length int) {
	if c.ID == viewmodel.StatusID {
		sb.WriteString(ui.Dark(utils.Text("", length)))
		return
	}
	sb.WriteString(ui.Dark(utils.Text(c.AuthorTime.Format(RFC3339Small), length)))
}

func writeSubject(sb *strings.Builder, c viewmodel.Commit, selectedCommit viewmodel.Commit, length int) {
	subject := utils.Text(c.Subject, length)
	if c.ID == viewmodel.StatusID {
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
