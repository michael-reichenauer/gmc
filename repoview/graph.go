package repoview

import (
	"github.com/michael-reichenauer/gmc/repoview/model"
	"github.com/michael-reichenauer/gmc/utils"
	"github.com/michael-reichenauer/gmc/utils/ui"
	"hash/fnv"
)

func hasLeft(bm utils.Bitmask) bool {
	return bm.Has(model.BBranchLeft) ||
		bm.Has(model.BPass) ||
		bm.Has(model.BMergeLeft)
}

func graphBranchRune(bm utils.Bitmask) rune {
	switch {
	// commit of a branch with only one commit (tip==bottom)
	case bm.Has(model.BTip) && bm.Has(model.BBottom) && bm.Has(model.BActiveTip) && hasLeft(bm):
		return '┺'
	case bm.Has(model.BTip) && bm.Has(model.BBottom) && hasLeft(bm):
		return '╼'

	// commit is tip
	case bm.Has(model.BTip) && bm.Has(model.BActiveTip) && hasLeft(bm):
		return '╊'
	case bm.Has(model.BTip) && bm.Has(model.BActiveTip):
		return '┣'
	case bm.Has(model.BTip) && hasLeft(bm):
		return '┲'
	case bm.Has(model.BTip):
		return '┏'

	// commit is bottom
	case bm.Has(model.BBottom) && hasLeft(bm):
		return '┺'
	case bm.Has(model.BBottom):
		return '┚'

	// commit is within branch
	case bm.Has(model.BCommit) && hasLeft(bm):
		return '╊'
	case bm.Has(model.BCommit):
		return '┣'

	// commit is not part of branch
	case bm.Has(model.BLine) && hasLeft(bm):
		return '╂'
	case bm.Has(model.BLine):
		return '┃'

	case bm == model.BPass:
		return '─'
	case bm == model.BBlank:
		return ' '
	default:
		return '*'
	}
}

//func graphBranchRune(bm utils.Bitmask) rune {
//	switch bm {
//	case model.BCommit:
//		return '┣'
//	case model.BCommit | model.BPass:
//		return '╊'
//	case model.BCommit | model.BMergeLeft:
//		return '╊'
//	case model.BCommit | model.BBranchLeft:
//		return '╊'
//	case model.BTip:
//		return '┏'
//	case model.BTip | model.BMergeLeft:
//		return '┺'
//	case model.BTip | model.BMergeLeft | model.BBranchLeft:
//		return '┲'
//	case model.BTip | model.BPass:
//		return '┏'
//	case model.BTip | model.BBranchLeft:
//		return '┲'
//	case model.BBottom:
//		return '┗'
//	case model.BBottom | model.BMergeLeft:
//		return '┺'
//	case model.BBottom | model.BPass:
//		return '┚'
//	case model.BLine:
//		return '┃'
//	case model.BLine | model.BPass:
//		return '╂'
//	case model.BPass:
//		return '─'
//	case model.BBlank:
//		return ' '
//	default:
//		return '*'
//	}
//}
func graphConnectRune(bm utils.Bitmask) rune {
	switch bm {
	case model.BMergeRight:
		return '╮'
	case model.BMergeRight | model.BPass:
		return '┬'
	case model.BMergeRight | model.BMLine:
		return '┤'
	case model.BMergeRight | model.BBranchRight:
		return '┤'
	case model.BMergeRight | model.BBranchRight | model.BPass:
		return '┴'
	case model.BBranchRight:
		return '╯'
	case model.BBranchRight | model.BPass:
		return '┴'
	case model.BBranchRight | model.BMLine:
		return '┤'
	case model.BMergeLeft:
		return '╭'
	case model.BMergeLeft | model.BBranchLeft:
		return '├'
	case model.BMergeLeft | model.BMLine:
		return '├'
	case model.BBranchLeft:
		return '╰'
	case model.BBranchLeft | model.BMLine:
		return '├'
	case model.BMLine:
		return '│'
	case model.BPass:
		return '─'
	case model.BBlank:
		return ' '
	default:
		return '*'
	}
}
func branchColor(name string, isMultiBranch bool) ui.Color {
	if name == "master:local" {
		return ui.CMagenta
	}
	if isMultiBranch {
		return ui.CWhite
	}
	h := fnv.New32a()
	h.Write([]byte(name))
	index := int(h.Sum32()) % len(branchColors)
	return branchColors[index]
}
