package repoview

import (
	"github.com/michael-reichenauer/gmc/repoview/viewmodel"
	"github.com/michael-reichenauer/gmc/utils"
	"github.com/michael-reichenauer/gmc/utils/ui"
)

type repoGraph struct {
}

var (
	currentCommitMarker = ui.White("●")
	mergeInMarker       = ui.Dark("╮")
	branchOurMarker     = ui.Dark("╯")
	inOutMarker         = ui.Dark("<")
)

func newRepoGraph() *repoGraph {
	return &repoGraph{}
}

func (t *repoGraph) graphBranchRune(bm utils.Bitmask) rune {
	switch {
	// commit of a branch with only one commit (tip==bottom)
	case bm.Has(viewmodel.BTip) && bm.Has(viewmodel.BBottom) && bm.Has(viewmodel.BActiveTip) && t.hasLeft(bm):
		return '┺'
	case bm.Has(viewmodel.BTip) && bm.Has(viewmodel.BBottom) && t.hasLeft(bm):
		return '╼'

	// commit is tip
	case bm.Has(viewmodel.BTip) && bm.Has(viewmodel.BActiveTip) && t.hasLeft(bm):
		return '╊'
	case bm.Has(viewmodel.BTip) && bm.Has(viewmodel.BActiveTip):
		return '┣'
	case bm.Has(viewmodel.BTip) && t.hasLeft(bm):
		return '┲'
	case bm.Has(viewmodel.BTip):
		return '┏'

	// commit is bottom
	case bm.Has(viewmodel.BBottom) && t.hasLeft(bm):
		return '┺'
	case bm.Has(viewmodel.BBottom):
		return '┚'

	// commit is within branch
	case bm.Has(viewmodel.BCommit) && t.hasLeft(bm):
		return '╊'
	case bm.Has(viewmodel.BCommit):
		return '┣'

	// commit is not part of branch
	case bm.Has(viewmodel.BLine) && t.hasLeft(bm):
		return '╂'
	case bm.Has(viewmodel.BLine):
		return '┃'

	case bm == viewmodel.BPass:
		return '─'
	case bm == viewmodel.BBlank:
		return ' '
	default:
		return '*'
	}
}

func (t *repoGraph) graphConnectRune(bm utils.Bitmask) rune {
	switch bm {
	case viewmodel.BMergeRight:
		return '╮'
	case viewmodel.BMergeRight | viewmodel.BPass:
		return '┬'
	case viewmodel.BMergeRight | viewmodel.BMLine:
		return '┤'
	case viewmodel.BMergeRight | viewmodel.BBranchRight:
		return '┤'
	case viewmodel.BMergeRight | viewmodel.BBranchRight | viewmodel.BPass:
		return '┴'
	case viewmodel.BBranchRight:
		return '╯'
	case viewmodel.BBranchRight | viewmodel.BMLine | viewmodel.BPass:
		return '┼'
	case viewmodel.BBranchRight | viewmodel.BPass:
		return '┴'
	case viewmodel.BBranchRight | viewmodel.BMLine:
		return '┤'
	case viewmodel.BMergeLeft:
		return '╭'
	case viewmodel.BMergeLeft | viewmodel.BBranchLeft:
		return '├'
	case viewmodel.BMergeLeft | viewmodel.BMLine:
		return '├'
	case viewmodel.BBranchLeft:
		return '╰'
	case viewmodel.BBranchLeft | viewmodel.BMLine:
		return '├'
	case viewmodel.BMLine | viewmodel.BPass:
		return '┼'
	case viewmodel.BMLine:
		return '│'
	case viewmodel.BPass:
		return '─'
	case viewmodel.BBlank:
		return ' '
	default:
		return '*'
	}
}

func (t *repoGraph) hasLeft(bm utils.Bitmask) bool {
	return bm.Has(viewmodel.BBranchLeft) ||
		bm.Has(viewmodel.BMergeLeft) ||
		bm.Has(viewmodel.BPass)
}
