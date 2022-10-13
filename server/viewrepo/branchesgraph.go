package viewrepo

import (
	"github.com/michael-reichenauer/gmc/api"
	"github.com/michael-reichenauer/gmc/utils/cui"
)

type BranchesGraph interface {
	SetGraph(repo *repo)
}

var branchColors = []cui.Color{
	cui.CMagenta,
	// cui.CMagentaDk,
	cui.CRed,
	// cui.CRedDk,
	cui.CBlue,
	// cui.CBlueDk,
	// cui.CYellow,
	cui.CYellowDk,
	cui.CGreen,
	// cui.CGreenDk,
	cui.CCyan,
	// cui.CCyanDk,
}

type branchesGraph struct {
}

func newBranchesGraph() BranchesGraph {
	return &branchesGraph{}
}

func (t *branchesGraph) SetGraph(repo *repo) {
	t.setBranchesXLocation(repo)

	for _, b := range repo.Branches {
		for y := b.tip.Index; y <= b.bottom.Index; y++ {
			c := repo.Commits[y]

			if c == b.tip && c.Branch != b {
				// this tip commit is not on this branch (multiple branch tips on the same commit)
				t.drawOtherBranchTip(repo, b, c)
				continue
			}

			t.drawBranch(repo, b, c) // Drawing either ┏  ┣ ┃ ┗

			if c.MergeParent != nil {
				t.drawMerge(repo, c) // Drawing   ╭ or  ╮
			}
			if c.Parent != nil && c.Parent.Branch != c.Branch {
				// Commit parent is on other branch (i.e. commit is first/bottom commit on this branch)
				// Draw branched from parent branch  ╯ or ╰
				t.drawBranchFromParent(repo, c)
			}
		}
	}
}

func (t *branchesGraph) drawOtherBranchTip(repo *repo, b *branch, c *commit) {
	x := b.x
	y := c.Index
	color := b.color
	// this tip commit is not part of the branch (multiple branch tips on the same commit)
	repo.drawHorizontalLine(c.Branch.x+1, x+1, y, color)   //              ─
	repo.SetGraphBranch(x, y, api.BBottom|api.Pass, color) //               ┺

}

func (t *branchesGraph) drawBranch(repo *repo, b *branch, c *commit) {
	x := b.x
	y := c.Index
	color := b.color

	if c.Branch != b && c != b.tip {
		// Other branch commit, normal branch line (no commit on that branch)
		repo.SetGraphBranch(x, y, api.BLine, color) // ┃
		return
	}

	if c.Branch != b {
		return
	}

	if c == c.Branch.tip {
		repo.SetGraphBranch(x, y, api.BTip, color) //       ┏   (branch tip)
	}
	if c == c.Branch.tip && c.Branch.isGitBranch {
		repo.SetGraphBranch(x, y, api.BActiveTip, color) // ┣   (indicate possible more commits in the future)
	}
	if c == c.Branch.bottom {
		repo.SetGraphBranch(x, y, api.BBottom, color) //    ┗   (bottom commit (e.g. initial commit on main)
	}
	if c != c.Branch.tip && c != c.Branch.bottom { //       ┣   (normal commit, in the middle)
		repo.SetGraphBranch(x, y, api.BCommit, color)
	}
}

func (t *branchesGraph) setBranchesXLocation(repo *repo) {
	for i, b := range repo.Branches {
		b.x = i
		// graphIndex := 0
		// for _, bb := range repo.Branches[:i] {
		// 	if bb.tip.Index <= b.tip.Index && bb.bottom.Index >= b.bottom.Index ||
		// 		bb.tip.Index <= b.tip.Index && bb.bottom.Index >= b.tip.Index ||
		// 		bb.tip.Index >= b.tip.Index && bb.tip.Index <= b.bottom.Index ||
		// 		bb.tip.Index >= b.tip.Index && bb.bottom.Index <= b.bottom.Index {
		// 		graphIndex++
		// 	}
		// }
		// b.graphIndex = graphIndex
	}
}

func (s *branchesGraph) drawMerge(repo *repo, commit *commit) {
	// Commit is a merge commit, has 2 parents
	if commit.MergeParent.Branch.index < commit.Branch.index {
		// Other branch is on the left side, merged from parent parent branch ╭
		s.drawMergeFromChildBranch(repo, commit)
	} else {
		// Other branch is on the right side, merged from child branch,  ╮
		s.drawMergeFromParentBranch(repo, commit)
	}
}

func (s *branchesGraph) drawMergeFromChildBranch(repo *repo, commit *commit) {
	x := commit.Branch.x
	y := commit.Index
	x2 := commit.MergeParent.Branch.x
	y2 := commit.MergeParent.Index

	// Other branch is on the left side, merged from parent parent branch ╭
	color := commit.Branch.color

	repo.SetGraphBranch(x, y, api.MergeFromLeft, color) //     ╭
	repo.SetGraphConnect(x, y, api.MergeFromLeft, color)
	if commit.Branch != commit.MergeParent.Branch {
		repo.drawVerticalLine(x, y+1, y2, color) //            │
	}
	repo.SetGraphConnect(x, y2, api.BranchToRight, color) //   ╯
	repo.drawHorizontalLine(x2+1, x, y2, color)           // ──
}

func (s *branchesGraph) drawMergeFromParentBranch(repo *repo, commit *commit) {
	// Commit is a merge commit, has 2 parents
	x := commit.Branch.x
	y := commit.Index
	x2 := commit.MergeParent.Branch.x
	y2 := commit.MergeParent.Index

	// Other branch is on the right side, merged from child branch,  ╮
	color := commit.MergeParent.Branch.color

	repo.drawHorizontalLine(x+1, x2, y, color) //                 ─
	if commit.Branch != commit.MergeParent.Branch {
		repo.SetGraphConnect(x2, y, api.MergeFromRight, color) //   ╮
	}
	if commit.Branch != commit.MergeParent.Branch {
		repo.drawVerticalLine(x2, y+1, y2, color) //                │
	}
	if commit.Branch != commit.MergeParent.Branch {
		repo.SetGraphBranch(x2, y2, api.BranchToLeft, color) //     ╰
		repo.SetGraphConnect(x2, y2, api.BranchToLeft, color)
	} else {
		repo.SetGraphBranch(x2, y2, api.BCommit, color) //          ┣
	}
}

func (s *branchesGraph) drawBranchFromParent(repo *repo, c *commit) {
	// Commit parent is on other branch (commit is first/bottom commit on this branch)
	// Branched from parent branch
	x := c.Branch.x
	y := c.Index
	x2 := c.Parent.Branch.x
	y2 := c.Parent.Index
	color := c.Branch.color

	if c.Parent.Branch.index < c.Branch.index {
		// Other branch is left side  ╭
		repo.SetGraphBranch(x, y, api.MergeFromLeft, color)
		repo.SetGraphConnect(x, y, api.MergeFromLeft, color)  //    ╭
		repo.drawVerticalLine(x, y+1, y2, color)              //    │
		repo.SetGraphConnect(x, y2, api.BranchToRight, color) //    ╯
		repo.drawHorizontalLine(x2+1, x, y2, color)           //  ──
	} else {
		// Other branch is right side, branched from some child branch ╮ (is this still valid ????)
		repo.SetGraphConnect(x+1, y, api.MergeFromRight, color) // ╮
		repo.drawVerticalLine(x+1, y+1, y2, color)              // │
		repo.SetGraphBranch(x2, y2, api.BranchToLeft, color)    // ╰
		repo.SetGraphConnect(x2, y2, api.BranchToLeft, color)
	}
}
