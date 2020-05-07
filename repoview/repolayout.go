package repoview

import (
	"fmt"
	"github.com/michael-reichenauer/gmc/repoview/viewmodel"
	"github.com/michael-reichenauer/gmc/utils"
	"github.com/michael-reichenauer/gmc/utils/ui"
	"strings"
)

const (
	dateTimeColumnFormat = "2006-01-02 15:04"
	markersWidth         = 4
)

type repoLayout struct {
	viewModelService *viewmodel.Service
	repoGraph        *repoGraph
}

func newRepoLayout(viewModelService *viewmodel.Service) *repoLayout {
	return &repoLayout{viewModelService: viewModelService, repoGraph: newRepoGraph()}
}

func (t *repoLayout) getPageLines(
	commits []viewmodel.Commit,
	viewWidth int,
	currentBranchDisplayName string,
	repo viewmodel.ViewRepo) []string {
	if len(commits) < 1 {
		return nil
	}

	graphWidth := t.getGraphWidth(commits)
	commitWidth := viewWidth - graphWidth
	messageWidth, authorWidth, timeWidth := t.columnWidths(commitWidth)

	var lines []string
	for _, c := range commits {
		var sb strings.Builder
		t.writeGraph(&sb, c)
		sb.WriteString(" ")
		t.writeMoreMarker(&sb, c)
		t.writeCurrentMarker(&sb, c)
		t.writeAheadBehindMarker(&sb, c)
		t.writeSubject(&sb, c, currentBranchDisplayName, messageWidth, repo)
		sb.WriteString(" ")
		t.writeAuthor(&sb, c, authorWidth)
		sb.WriteString(" ")
		t.writeAuthorTime(&sb, c, timeWidth)

		lines = append(lines, sb.String())
	}
	return lines
}

func (t *repoLayout) getMoreIndex(repo viewmodel.ViewRepo) int {
	graphWidth := t.getGraphWidth(repo.Commits)
	return graphWidth - 2
}

func (t *repoLayout) getGraphWidth(commits []viewmodel.Commit) int {
	if len(commits) == 0 {
		return 0
	}
	return len(commits[0].Graph)*2 + markersWidth
}

func (t *repoLayout) columnWidths(commitWidth int) (msgLength, authorWidth, timeWidth int) {
	// Default widths (norma and wide view)
	authorWidth = 15
	timeWidth = 16

	if commitWidth < 60 {
		// Disabled author and time if very narrow view
		authorWidth = 0
		timeWidth = 0
	} else if commitWidth < 90 {
		// Reducing author and and time if narrow view
		authorWidth = 10
		timeWidth = 8
	}

	msgLength = commitWidth - authorWidth - timeWidth
	if msgLength < 0 {
		msgLength = 0
	}
	return
}

func (t *repoLayout) writeGraph(sb *strings.Builder, c viewmodel.Commit) {
	for i := 0; i < len(c.Graph); i++ {
		// Normal branch color
		bColor := t.viewModelService.BranchColor(c.Graph[i].BranchDisplayName)
		//	if i != 0 {
		cColor := bColor
		if c.Graph[i].Connect == viewmodel.BPass &&
			c.Graph[i].PassName != "" &&
			c.Graph[i].PassName != "-" {
			cColor = t.viewModelService.BranchColor(c.Graph[i].PassName)
		} else if c.Graph[i].Connect.Has(viewmodel.BPass) {
			cColor = ui.CWhite
		}
		sb.WriteString(ui.ColorRune(cColor, t.repoGraph.graphConnectRune(c.Graph[i].Connect)))
		//	}

		if c.Graph[i].Branch == viewmodel.BPass &&
			c.Graph[i].PassName != "" &&
			c.Graph[i].PassName != "-" {
			bColor = t.viewModelService.BranchColor(c.Graph[i].PassName)
		} else if c.Graph[i].Branch == viewmodel.BPass {
			bColor = ui.CWhite
		}

		sb.WriteString(ui.ColorRune(bColor, t.repoGraph.graphBranchRune(c.Graph[i].Branch)))
	}
}

func (t *repoLayout) writeCurrentMarker(sb *strings.Builder, c viewmodel.Commit) {
	if c.IsCurrent {
		sb.WriteString(currentCommitMarker)
	} else {
		sb.WriteString(" ")
	}
}

func (t *repoLayout) writeMoreMarker(sb *strings.Builder, c viewmodel.Commit) {
	if c.More.Has(viewmodel.MoreMergeIn) && c.More.Has(viewmodel.MoreBranchOut) {
		sb.WriteString(inOutMarker)
	} else if c.More.Has(viewmodel.MoreMergeIn) {
		sb.WriteString(mergeInMarker)
	} else if c.More.Has(viewmodel.MoreBranchOut) {
		sb.WriteString(branchOurMarker)
	} else {
		sb.WriteString(" ")
	}
}

func (t *repoLayout) writeAheadBehindMarker(sb *strings.Builder, c viewmodel.Commit) {
	if c.IsLocalOnly {
		sb.WriteString(ui.GreenDk("▲"))
	} else if c.IsRemoteOnly {
		sb.WriteString(ui.Blue("▼"))
	} else {
		sb.WriteString(" ")
	}
}

func (t *repoLayout) writeAuthor(sb *strings.Builder, commit viewmodel.Commit, length int) {
	sb.WriteString(ui.Dark(utils.Text(commit.Author, length)))
}

func (t *repoLayout) writeAuthorTime(sb *strings.Builder, c viewmodel.Commit, length int) {
	if c.ID == viewmodel.UncommittedID {
		sb.WriteString(ui.Dark(utils.Text("", length)))
		return
	}
	tt := c.AuthorTime.Format(dateTimeColumnFormat)
	//tt = strings.Replace(tt, "-", "", -1)

	tt = tt[2:]
	sb.WriteString(ui.Dark(utils.Text(tt, length)))
}

func (t *repoLayout) writeSubject(sb *strings.Builder, c viewmodel.Commit, currentBranchDisplayName string, length int, repo viewmodel.ViewRepo) {
	tagsText := t.toTagsText(c, length)

	subject := utils.Text(c.Subject, length-len(tagsText))
	if c.ID == viewmodel.PartialLogCommitID {
		sb.WriteString(ui.Dark(subject))
		return
	}
	if c.ID == viewmodel.UncommittedID {
		if repo.Conflicts > 0 {
			sb.WriteString(ui.Red(subject))
			return
		}
		if repo.MergeMessage != "" {
			sb.WriteString(ui.RedDk(subject))
			return
		}
		sb.WriteString(ui.YellowDk(subject))
		return
	}
	color := ui.CWhite
	if c.IsLocalOnly {
		color = ui.CGreenDk
	} else if c.IsRemoteOnly {
		color = ui.CBlue
	}
	if currentBranchDisplayName != "" &&
		c.Branch.DisplayName != currentBranchDisplayName {
		color = ui.CDark
	}
	sb.WriteString(ui.Green(tagsText))
	sb.WriteString(ui.ColorText(color, subject))
}

func (t *repoLayout) toTagsText(c viewmodel.Commit, lenght int) string {
	if len(c.Tags) == 0 {
		return ""
	}
	text := fmt.Sprintf("%v", c.Tags)
	if len(text) > lenght/2 {
		text = text[:lenght/2] + "...]"
	}
	return text + " "
}
