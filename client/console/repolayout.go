package console

import (
	"fmt"
	"github.com/michael-reichenauer/gmc/api"
	"github.com/michael-reichenauer/gmc/utils"
	"github.com/michael-reichenauer/gmc/utils/cui"
	"strings"
)

const (
	dateTimeColumnFormat = "2006-01-02 15:04"
	markersWidth         = 4
)

type repoLayout struct {
	repoGraph *repoGraph
}

func newRepoLayout() *repoLayout {
	return &repoLayout{repoGraph: newRepoGraph()}
}

func (t *repoLayout) getPageLines(
	commits []api.Commit,
	graphRows []api.GraphRow,
	viewWidth int,
	currentBranchDisplayName string,
	repo api.ViewRepo,
) []string {
	if len(commits) < 1 {
		return nil
	}

	graphWidth := t.getGraphWidth(graphRows)
	commitWidth := viewWidth - graphWidth
	messageWidth, authorWidth, timeWidth := t.columnWidths(commitWidth)

	var lines []string
	for i, c := range commits {
		var sb strings.Builder
		t.writeGraph(&sb, graphRows[i])
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

func (t *repoLayout) getMoreIndex(repo api.ViewRepo) int {
	graphWidth := t.getGraphWidth(repo.ConsoleGraph)
	return graphWidth - 2
}

func (t *repoLayout) getGraphWidth(graph api.Graph) int {
	if len(graph) == 0 {
		return 0
	}
	//return markersWidth
	return len(graph[0])*2 + markersWidth
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

func (t *repoLayout) writeGraph(sb *strings.Builder, row api.GraphRow) {
	for i := 0; i < len(row); i++ {
		// Normal branch color
		bColor := cui.Color(row[i].BranchColor) //t.branchColors.BranchColor(c.Graph[i].BranchDisplayName)

		cColor := bColor
		if row[i].Connect == api.BPass &&
			row[i].PassColor != 0 &&
			row[i].PassColor != api.Color(cui.CWhite) {
			cColor = cui.Color(row[i].PassColor) // t.branchColors.BranchColor(c.Graph[i].PassName)
		} else if row[i].Connect.Has(api.BPass) {
			cColor = cui.CWhite
		}
		sb.WriteString(cui.ColorRune(cColor, t.repoGraph.graphConnectRune(row[i].Connect)))

		if row[i].Branch == api.BPass &&
			row[i].PassColor != 0 &&
			row[i].PassColor != api.Color(cui.CWhite) {
			bColor = cui.Color(row[i].PassColor) // t.branchColors.BranchColor(c.Graph[i].PassName)
		} else if row[i].Branch == api.BPass {
			bColor = cui.CWhite
		}

		sb.WriteString(cui.ColorRune(bColor, t.repoGraph.graphBranchRune(row[i].Branch)))
	}
}

func (t *repoLayout) writeCurrentMarker(sb *strings.Builder, c api.Commit) {
	if c.IsCurrent {
		sb.WriteString(currentCommitMarker)
	} else {
		sb.WriteString(" ")
	}
}

func (t *repoLayout) writeMoreMarker(sb *strings.Builder, c api.Commit) {
	if c.More.Has(api.MoreMergeIn) && c.More.Has(api.MoreBranchOut) {
		sb.WriteString(inOutMarker)
	} else if c.More.Has(api.MoreMergeIn) {
		sb.WriteString(mergeInMarker)
	} else if c.More.Has(api.MoreBranchOut) {
		sb.WriteString(branchOurMarker)
	} else {
		sb.WriteString(" ")
	}
}

func (t *repoLayout) writeAheadBehindMarker(sb *strings.Builder, c api.Commit) {
	if c.IsLocalOnly {
		sb.WriteString(cui.GreenDk("▲"))
	} else if c.IsRemoteOnly {
		sb.WriteString(cui.Blue("▼"))
	} else {
		sb.WriteString(" ")
	}
}

func (t *repoLayout) writeAuthor(sb *strings.Builder, commit api.Commit, length int) {
	sb.WriteString(cui.Dark(utils.Text(commit.Author, length)))
}

func (t *repoLayout) writeAuthorTime(sb *strings.Builder, c api.Commit, length int) {
	if c.IsUncommitted {
		sb.WriteString(cui.Dark(utils.Text("", length)))
		return
	}
	tt := c.AuthorTime.Format(dateTimeColumnFormat)
	//tt = strings.Replace(tt, "-", "", -1)

	tt = tt[2:]
	sb.WriteString(cui.Dark(utils.Text(tt, length)))
}

func (t *repoLayout) writeSubject(
	sb *strings.Builder,
	c api.Commit,
	currentBranchDisplayName string,
	length int,
	repo api.ViewRepo,
) {
	tagsText := t.toTagsText(c, length)

	subject := utils.Text(c.Subject, length-len(tagsText))
	if c.IsPartialLogCommit {
		sb.WriteString(cui.Dark(subject))
		return
	}
	if c.IsUncommitted {
		if repo.Conflicts > 0 {
			sb.WriteString(cui.Red(subject))
			return
		}
		if repo.MergeMessage != "" {
			sb.WriteString(cui.RedDk(subject))
			return
		}
		sb.WriteString(cui.YellowDk(subject))
		return
	}
	color := cui.CWhite
	if c.IsLocalOnly {
		color = cui.CGreenDk
	} else if c.IsRemoteOnly {
		color = cui.CBlue
	}
	if currentBranchDisplayName != "" &&
		repo.Branches[c.BranchIndex].DisplayName != currentBranchDisplayName {
		color = cui.CDark
	}
	sb.WriteString(cui.Green(tagsText))
	sb.WriteString(cui.ColorText(color, subject))
}

func (t *repoLayout) toTagsText(c api.Commit, length int) string {
	if len(c.Tags) == 0 {
		return ""
	}
	text := fmt.Sprintf("%v", c.Tags)
	if len(text) > length/2 {
		text = text[:length/2] + "...]"
	}
	return text + " "
}
