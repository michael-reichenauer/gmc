package console

import (
	"fmt"
	"strings"

	"github.com/michael-reichenauer/gmc/api"
	"github.com/michael-reichenauer/gmc/utils"
	"github.com/michael-reichenauer/gmc/utils/cui"
)

const (
	dateTimeColumnFormat = "2006-01-02 15:04"
	markersWidth         = 2
)

var (
	currentCommitMarker = cui.White("●")
)

type repoLayout struct {
	repoGraph *RepoGraph
}

func newRepoLayout() *repoLayout {
	return &repoLayout{repoGraph: NewRepoGraph()}
}

type tip struct {
	len  int
	text string
}

func (t *repoLayout) getPageLines(
	commits []api.Commit,
	graphRows []api.GraphRow,
	viewWidth int,
	selectedBranchName string,
	repo api.Repo,
) []string {
	if len(commits) < 1 {
		return nil
	}

	tips := t.getBranchTips(repo)

	graphWidth := t.getGraphWidth(graphRows)
	commitWidth := viewWidth - graphWidth
	messageWidth, sidWidth, authorWidth, timeWidth := t.columnWidths(commitWidth)

	var lines []string
	for i, c := range commits {
		var sb strings.Builder
		t.writeGraph(&sb, graphRows[i])
		t.writeCurrentMarker(&sb, c)
		t.writeAheadBehindMarker(&sb, c)
		t.writeSubject(&sb, c, selectedBranchName, messageWidth, repo, tips)
		t.writeSid(&sb, c, sidWidth)
		t.writeAuthor(&sb, c, authorWidth)
		t.writeAuthorTime(&sb, c, timeWidth)

		lines = append(lines, sb.String())
	}
	return lines
}

func (t *repoLayout) getBranchTips(repo api.Repo) map[string]tip {
	tm := make(map[string]tip)
	for _, b := range repo.Branches {

		t, ok := tm[b.TipID]
		if !ok {
			t = tip{}
		}

		if b.AmbiguousTipId != "" {
			at, ok := tm[b.AmbiguousTipId]
			if !ok {
				at = tip{}
			}

			atx := "ambiguous"
			at.text = at.text + cui.White("(╸") + cui.Dark(atx) + cui.White(")")
			at.len = at.len + len(atx) + 3
			tm[b.AmbiguousTipId] = at
		}

		if t.len > 40 {
			// To many tips
			if !strings.HasSuffix(t.text, "(...") {
				t.text = t.text + cui.Dark("(...")
				t.len = t.len + 3
				tm[b.TipID] = t
			}
			continue
		}
		txt := b.DisplayName
		if len(txt) > 16 {
			txt = "..." + txt[len(txt)-16:]
		}
		if b.IsRemote {
			txt = "^/" + txt
		}

		t.len = t.len + len(txt) + 2
		color := cui.Color(b.Color)

		tagTxt := cui.ColorText(color, "("+txt+")")
		if !b.IsGitBranch {
			tagTxt = cui.ColorText(color, "(╸") + cui.Dark(txt) + cui.ColorText(color, ")")
			t.len = t.len + 1
		}

		t.text = t.text + tagTxt
		tm[b.TipID] = t
	}

	return tm
}

func (t *repoLayout) getSubjectXCoordinate(repo api.Repo) int {
	graphWidth := t.getGraphWidth(repo.ConsoleGraph)
	return graphWidth
}

func (t *repoLayout) getGraphWidth(graph api.Graph) int {
	if len(graph) == 0 {
		return 0
	}
	//return markersWidth
	return len(graph[0])*2 + markersWidth
}

func (t *repoLayout) columnWidths(commitWidth int) (msgLength, sidWidth, authorWidth, timeWidth int) {
	// Default widths (norma and wide view)
	authorWidth = 15
	timeWidth = 16
	sidWidth = 6
	spaceWidth := 1

	if commitWidth < 60 {
		// Disabled author and time if very narrow view
		sidWidth = 0
		authorWidth = 0
		timeWidth = 0
	} else if commitWidth < 100 {
		// Reducing author and and time if narrow view
		sidWidth = 0
		authorWidth = 10
		timeWidth = 8
		spaceWidth = 2
	}

	msgLength = commitWidth - sidWidth - authorWidth - timeWidth - spaceWidth
	if msgLength < 0 {
		msgLength = 0
	}
	return
}

func (t *repoLayout) writeGraph(sb *strings.Builder, row api.GraphRow) {
	t.repoGraph.WriteGraph(sb, row)
}

func (t *repoLayout) writeCurrentMarker(sb *strings.Builder, c api.Commit) {
	if c.IsCurrent {
		sb.WriteString(currentCommitMarker)
	} else {
		sb.WriteString(" ")
	}
}

// func (t *repoLayout) writeMoreMarker(sb *strings.Builder, c api.Commit) {
// 	if c.More.Has(api.MoreMergeIn) && c.More.Has(api.MoreBranchOut) {
// 		sb.WriteString(inOutMarker)
// 	} else if c.More.Has(api.MoreMergeIn) {
// 		sb.WriteString(mergeInMarker)
// 	} else if c.More.Has(api.MoreBranchOut) {
// 		sb.WriteString(branchOutMarker)
// 	} else {
// 		sb.WriteString(" ")
// 	}
// }

func (t *repoLayout) writeAheadBehindMarker(sb *strings.Builder, c api.Commit) {
	if c.IsLocalOnly {
		sb.WriteString(cui.GreenDk("▲"))
	} else if c.IsRemoteOnly {
		sb.WriteString(cui.Blue("▼"))
	} else {
		sb.WriteString(" ")
	}
}

func (t *repoLayout) writeSid(sb *strings.Builder, commit api.Commit, length int) {
	if length > 0 {
		sb.WriteString(" ")
	}
	sb.WriteString(cui.Dark(utils.Text(commit.SID, length)))
}

func (t *repoLayout) writeAuthor(sb *strings.Builder, commit api.Commit, length int) {
	if length > 0 {
		sb.WriteString(" ")
	}
	sb.WriteString(cui.Dark(utils.Text(commit.Author, length)))
}

func (t *repoLayout) writeAuthorTime(sb *strings.Builder, c api.Commit, length int) {
	if length > 0 {
		sb.WriteString(" ")
	}
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
	repo api.Repo,
	tips map[string]tip,
) {
	tagsText := t.toTagsText(c, length)
	tipsText := ""
	tipsLen := 0
	tip, ok := tips[c.ID]
	if ok {
		tipsText = tip.text + " "
		tipsLen = tip.len + 1
	}

	subject := utils.Text(c.Subject, length-len(tagsText)-tipsLen)
	if c.IsPartialLogCommit {
		sb.WriteString(cui.Dark(subject))
		return
	}
	if c.IsUncommitted {
		if repo.Conflicts > 0 {
			sb.WriteString(cui.Red(subject))
			return
		}

		sb.WriteString(tipsText)
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
	sb.WriteString(tipsText)
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
