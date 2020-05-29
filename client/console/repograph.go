package console

import (
	"github.com/michael-reichenauer/gmc/api"
	"github.com/michael-reichenauer/gmc/utils"
	"github.com/michael-reichenauer/gmc/utils/cui"
)

type repoGraph struct {
}

var (
	currentCommitMarker = cui.White("●")
	mergeInMarker       = cui.Dark("╮")
	branchOurMarker     = cui.Dark("╭")
	inOutMarker         = cui.Dark("<")
)

func newRepoGraph() *repoGraph {
	return &repoGraph{}
}

func (t *repoGraph) graphBranchRune(bm utils.Bitmask) rune {
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

	case bm == api.BPass:
		return '─'
	case bm == api.BBlank:
		return ' '
	default:
		return '*'
	}
}

func (t *repoGraph) graphConnectRune(bm utils.Bitmask) rune {
	switch bm {
	case api.BMergeRight:
		return '╮'
	case api.BMergeRight | api.BPass:
		return '┬'
	case api.BMergeRight | api.BMLine:
		return '┤'
	case api.BMergeRight | api.BBranchRight:
		return '┤'
	case api.BMergeRight | api.BBranchRight | api.BPass:
		return '┴'
	case api.BBranchRight:
		return '╯'
	case api.BBranchRight | api.BMLine | api.BPass:
		return '┼'
	case api.BBranchRight | api.BPass:
		return '┴'
	case api.BBranchRight | api.BMLine:
		return '┤'
	case api.BMergeLeft:
		return '╭'
	case api.BMergeLeft | api.BBranchLeft:
		return '├'
	case api.BMergeLeft | api.BMLine:
		return '├'
	case api.BBranchLeft:
		return '╰'
	case api.BBranchLeft | api.BMLine:
		return '├'
	case api.BMLine | api.BPass:
		return '┼'
	case api.BMLine:
		return '│'
	case api.BPass:
		return '─'
	case api.BBlank:
		return ' '
	default:
		return '*'
	}
}

func (t *repoGraph) hasLeft(bm utils.Bitmask) bool {
	return bm.Has(api.BBranchLeft) ||
		bm.Has(api.BMergeLeft) ||
		bm.Has(api.BPass)
}
