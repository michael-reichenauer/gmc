package console

import (
	"strings"

	"github.com/michael-reichenauer/gmc/api"
	"github.com/michael-reichenauer/gmc/utils"
	"github.com/michael-reichenauer/gmc/utils/cui"
)

type RepoGraph struct {
}

func NewRepoGraph() *RepoGraph {
	return &RepoGraph{}
}

func (t *RepoGraph) WriteGraph(sb *strings.Builder, row api.GraphRow) {
	for i := 0; i < len(row); i++ {
		// Colors
		branchColor := cui.Color(row[i].BranchColor)
		connectColor := cui.Color(row[i].ConnectColor)
		passColor := cui.Color(row[i].PassColor)

		// Draw connect runes (left of the branch)
		if row[i].Connect == api.Pass &&
			passColor != 0 &&
			passColor != cui.CWhite {
			connectColor = passColor
		} else if row[i].Connect.Has(api.Pass) {
			connectColor = cui.CWhite
		}
		sb.WriteString(cui.ColorRune(connectColor, t.graphConnectRune(row[i].Connect)))

		// Draw the branch rune
		if row[i].Branch == api.Pass &&
			passColor != 0 &&
			passColor != cui.CWhite {
			branchColor = passColor
		} else if row[i].Branch == api.Pass {
			branchColor = cui.CWhite
		}
		sb.WriteString(cui.ColorRune(branchColor, t.graphBranchRune(row[i].Branch)))
	}
}

func (t *RepoGraph) graphBranchRune(bm utils.Bitmask) rune {
	switch {
	// commit of a branch with only one commit (tip==bottom)
	case bm.Has(api.BTip) && bm.Has(api.BBottom) && bm.Has(api.BActiveTip) && t.hasLeft(bm):
		return '┺'
	case bm.Has(api.BTip) && bm.Has(api.BBottom) && t.hasLeft(bm):
		return '╼'

	// commit is tip
	case bm.Has(api.BTip) && bm.Has(api.BActiveTip) && t.hasLeft(bm):
		return '╊'
	case bm.Has(api.BTip) && bm.Has(api.BActiveTip):
		return '┣'
	case bm.Has(api.BTip) && t.hasLeft(bm):
		return '┲'
	case bm.Has(api.BTip):
		return '┏'

	// commit is bottom
	case bm.Has(api.BBottom) && t.hasLeft(bm):
		return '┺'
	case bm.Has(api.BBottom):
		return '┚'

	// commit is within branch
	case bm.Has(api.BCommit) && t.hasLeft(bm):
		return '╊'
	case bm.Has(api.BCommit):
		return '┣'

	// commit is not part of branch
	case bm.Has(api.BLine) && t.hasLeft(bm):
		return '╂'
	case bm.Has(api.BLine):
		return '┃'

	case bm == api.Pass:
		return '─'
	case bm == api.BBlank:
		return ' '
	default:
		return '*'
	}
}

func (t *RepoGraph) graphConnectRune(bm utils.Bitmask) rune {
	switch bm {
	case api.MergeFromRight:
		return '╮'
	case api.MergeFromRight | api.Pass:
		return '┬'
	case api.MergeFromRight | api.ConnectLine:
		return '┤'
	case api.MergeFromRight | api.BranchToRight:
		return '┤'
	case api.MergeFromRight | api.BranchToRight | api.Pass:
		return '┴'
	case api.BranchToRight:
		return '╯'
	case api.BranchToRight | api.ConnectLine | api.Pass:
		return '┼'
	case api.BranchToRight | api.Pass:
		return '┴'
	case api.BranchToRight | api.ConnectLine:
		return '┤'
	case api.MergeFromLeft:
		return '╭'
	case api.MergeFromLeft | api.BranchToLeft:
		return '├'
	case api.MergeFromLeft | api.ConnectLine:
		return '├'
	case api.BranchToLeft:
		return '╰'
	case api.BranchToLeft | api.ConnectLine:
		return '├'
	case api.ConnectLine | api.Pass:
		return '┼'
	case api.ConnectLine:
		return '│'
	case api.Pass:
		return '─'
	case api.BBlank:
		return ' '
	default:
		return '*'
	}
}

func (t *RepoGraph) hasLeft(bm utils.Bitmask) bool {
	return bm.Has(api.BranchToLeft) ||
		bm.Has(api.MergeFromLeft) ||
		bm.Has(api.Pass)
}
