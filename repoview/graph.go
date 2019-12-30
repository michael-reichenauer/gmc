package repoview

import (
	"github.com/michael-reichenauer/gmc/repoview/viewmodel"
	"github.com/michael-reichenauer/gmc/utils"
	"github.com/michael-reichenauer/gmc/utils/ui"
	"hash/fnv"
	"strings"
)

var (
	currentCommitMarker = ui.White("●")
	moreMarker          = ui.Dark(">")
)

var branchColors = []ui.Color{
	ui.CRed,
	ui.CBlue,
	ui.CYellow,
	ui.CGreen,
	ui.CCyan,
	ui.CRedDk,
	ui.CGreenDk,
	ui.CYellowDk,
	//ui.CBlueDk,
	ui.CMagentaDk,
	ui.CCyanDk,
}

func hasLeft(bm utils.Bitmask) bool {
	return bm.Has(viewmodel.BBranchLeft) ||
		bm.Has(viewmodel.BMergeLeft) ||
		bm.Has(viewmodel.BPass)
}

func graphBranchRune(bm utils.Bitmask) rune {
	switch {
	// commit of a branch with only one commit (tip==bottom)
	case bm.Has(viewmodel.BTip) && bm.Has(viewmodel.BBottom) && bm.Has(viewmodel.BActiveTip) && hasLeft(bm):
		return '┺'
	case bm.Has(viewmodel.BTip) && bm.Has(viewmodel.BBottom) && hasLeft(bm):
		return '╼'

	// commit is tip
	case bm.Has(viewmodel.BTip) && bm.Has(viewmodel.BActiveTip) && hasLeft(bm):
		return '╊'
	case bm.Has(viewmodel.BTip) && bm.Has(viewmodel.BActiveTip):
		return '┣'
	case bm.Has(viewmodel.BTip) && hasLeft(bm):
		return '┲'
	case bm.Has(viewmodel.BTip):
		return '┏'

	// commit is bottom
	case bm.Has(viewmodel.BBottom) && hasLeft(bm):
		return '┺'
	case bm.Has(viewmodel.BBottom):
		return '┚'

	// commit is within branch
	case bm.Has(viewmodel.BCommit) && hasLeft(bm):
		return '╊'
	case bm.Has(viewmodel.BCommit):
		return '┣'

	// commit is not part of branch
	case bm.Has(viewmodel.BLine) && hasLeft(bm):
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

func graphConnectRune(bm utils.Bitmask) rune {
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

func branchColor(name string) ui.Color {
	if name == "master" {
		return ui.CMagenta
	}
	if strings.HasPrefix(name, "multi:") {
		return ui.CWhite
	}
	h := fnv.New32a()
	h.Write([]byte(name))
	index := int(h.Sum32()) % len(branchColors)
	return branchColors[index]
}
