package repoview

import (
	"github.com/michael-reichenauer/gmc/utils"
	"github.com/michael-reichenauer/gmc/utils/cui"
	"github.com/michael-reichenauer/gmc/viewrepo"
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
	case bm.Has(viewrepo.BTip) && bm.Has(viewrepo.BBottom) && bm.Has(viewrepo.BActiveTip) && t.hasLeft(bm):
		return '┺'
	case bm.Has(viewrepo.BTip) && bm.Has(viewrepo.BBottom) && t.hasLeft(bm):
		return '╼'

	// commit is tip
	case bm.Has(viewrepo.BTip) && bm.Has(viewrepo.BActiveTip) && t.hasLeft(bm):
		return '╊'
	case bm.Has(viewrepo.BTip) && bm.Has(viewrepo.BActiveTip):
		return '┣'
	case bm.Has(viewrepo.BTip) && t.hasLeft(bm):
		return '┲'
	case bm.Has(viewrepo.BTip):
		return '┏'

	// commit is bottom
	case bm.Has(viewrepo.BBottom) && t.hasLeft(bm):
		return '┺'
	case bm.Has(viewrepo.BBottom):
		return '┚'

	// commit is within branch
	case bm.Has(viewrepo.BCommit) && t.hasLeft(bm):
		return '╊'
	case bm.Has(viewrepo.BCommit):
		return '┣'

	// commit is not part of branch
	case bm.Has(viewrepo.BLine) && t.hasLeft(bm):
		return '╂'
	case bm.Has(viewrepo.BLine):
		return '┃'

	case bm == viewrepo.BPass:
		return '─'
	case bm == viewrepo.BBlank:
		return ' '
	default:
		return '*'
	}
}

func (t *repoGraph) graphConnectRune(bm utils.Bitmask) rune {
	switch bm {
	case viewrepo.BMergeRight:
		return '╮'
	case viewrepo.BMergeRight | viewrepo.BPass:
		return '┬'
	case viewrepo.BMergeRight | viewrepo.BMLine:
		return '┤'
	case viewrepo.BMergeRight | viewrepo.BBranchRight:
		return '┤'
	case viewrepo.BMergeRight | viewrepo.BBranchRight | viewrepo.BPass:
		return '┴'
	case viewrepo.BBranchRight:
		return '╯'
	case viewrepo.BBranchRight | viewrepo.BMLine | viewrepo.BPass:
		return '┼'
	case viewrepo.BBranchRight | viewrepo.BPass:
		return '┴'
	case viewrepo.BBranchRight | viewrepo.BMLine:
		return '┤'
	case viewrepo.BMergeLeft:
		return '╭'
	case viewrepo.BMergeLeft | viewrepo.BBranchLeft:
		return '├'
	case viewrepo.BMergeLeft | viewrepo.BMLine:
		return '├'
	case viewrepo.BBranchLeft:
		return '╰'
	case viewrepo.BBranchLeft | viewrepo.BMLine:
		return '├'
	case viewrepo.BMLine | viewrepo.BPass:
		return '┼'
	case viewrepo.BMLine:
		return '│'
	case viewrepo.BPass:
		return '─'
	case viewrepo.BBlank:
		return ' '
	default:
		return '*'
	}
}

func (t *repoGraph) hasLeft(bm utils.Bitmask) bool {
	return bm.Has(viewrepo.BBranchLeft) ||
		bm.Has(viewrepo.BMergeLeft) ||
		bm.Has(viewrepo.BPass)
}
