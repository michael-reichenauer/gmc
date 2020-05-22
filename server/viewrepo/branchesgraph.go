package viewrepo

import (
	"github.com/michael-reichenauer/gmc/api"
	"github.com/michael-reichenauer/gmc/utils/cui"
)

var branchColors = []cui.Color{
	cui.CRed,
	cui.CBlue,
	cui.CYellow,
	cui.CGreen,
	cui.CCyan,
	cui.CRedDk,
	cui.CGreenDk,
	cui.CYellowDk,
	//ui.CBlueDk,
	cui.CMagenta,
	cui.CMagentaDk,
	cui.CCyanDk,
}

type branchesGraph struct {
}

func newBranchesGraph() *branchesGraph {
	return &branchesGraph{}
}

func (s *branchesGraph) drawBranchLines(repo *viewRepo) {
	for _, branch := range repo.Branches {
		branch.tip = repo.commitById[branch.tipId]
		s.drawBranchLine(branch)
	}
}

func (s *branchesGraph) drawBranchLine(branch *branch) {
	c := branch.tip
	for {
		if c.Branch != branch {
			// this commit is not part of the branch (multiple branches on the same commit)
			break
		}
		if c == c.Branch.tip {
			c.graph[branch.index].Branch.Set(api.BTip) //       ┏   (branch tip)
		}
		if c == c.Branch.tip && c.Branch.isGitBranch {
			c.graph[branch.index].Branch.Set(api.BActiveTip) // ┣   (indicate possible more commits in the future)
		}
		if c == c.Branch.bottom {
			c.graph[branch.index].Branch.Set(api.BBottom) // ┗   (bottom commit (e.g. initial commit on master)
		}
		if c != c.Branch.tip && c != c.Branch.bottom { // ┣   (normal commit, in the middle)
			c.graph[branch.index].Branch.Set(api.BCommit)
		}

		if c.Parent == nil || c.Branch != c.Parent.Branch {
			// Reached bottom of branch
			break
		}
		c = c.Parent
	}
}

func (s *branchesGraph) drawConnectorLines(repo *viewRepo) {
	for _, c := range repo.Commits {
		for i, b := range repo.Branches {
			//	c.graph[i].BranchName = b.name
			c.graph[i].BranchColor = api.Color(b.color)
			if c.Branch == b {
				// Commit branch
				if c.MergeParent != nil {
					// Commit has merge (2 parents)
					if c.MergeParent.Branch.index < c.Branch.index {
						// Other branch is left side ╭
						c.graph[i].Connect.Set(api.BMergeLeft)
						c.graph[i].Branch.Set(api.BMergeLeft)
						// Draw horizontal pass through line ───
						for k := c.MergeParent.Branch.index + 1; k < c.Branch.index; k++ {
							c.MergeParent.graph[k].Connect.Set(api.BPass)
							c.MergeParent.graph[k].Branch.Set(api.BPass)
							if c.MergeParent.graph[k].PassColor == 0 {
								c.MergeParent.graph[k].PassColor = api.Color(b.color)
							} else {
								c.MergeParent.graph[k].PassColor = api.Color(cui.CWhite)
							}
						}
						// Draw vertical down line │
						for j := c.Index + 1; j < c.MergeParent.Index; j++ {
							cc := repo.Commits[j]
							cc.graph[i].Connect.Set(api.BMLine)
						}
						// Draw ╯
						c.MergeParent.graph[i].Connect.Set(api.BBranchRight)
					} else {
						// Other branch is right side  ╮
						// Draw merge in rune
						c.graph[c.MergeParent.Branch.index].Connect.Set(api.BMergeRight)
						// Draw horizontal pass through line ───
						for k := i + 1; k < c.MergeParent.Branch.index; k++ {
							c.graph[k].Connect.Set(api.BPass)
							c.graph[k].Branch.Set(api.BPass)
							if c.graph[k].PassColor == 0 {
								c.graph[k].PassColor = api.Color(c.MergeParent.Branch.color)
							} else {
								c.graph[k].PassColor = api.Color(cui.CWhite)
							}
						}
						// Draw vertical down line │
						for j := c.Index + 1; j < c.MergeParent.Index; j++ {
							cc := repo.Commits[j]
							cc.graph[c.MergeParent.Branch.index].Connect.Set(api.BMLine)
						}
						// Draw branch out rune ╰
						c.MergeParent.graph[c.MergeParent.Branch.index].Connect.Set(api.BBranchLeft)
						c.MergeParent.graph[c.MergeParent.Branch.index].Branch.Set(api.BBranchLeft)
					}
				}
				if c.Parent != nil && c.Parent.Branch != c.Branch {
					// Commit parent is on other branch (bottom/first commit on this branch)
					if c.Parent.Branch.index < c.Branch.index {
						// Other branch is left side  ╭
						c.graph[i].Connect.Set(api.BMergeLeft)
						c.graph[i].Branch.Set(api.BMergeLeft)

						// ──
						for k := c.Parent.Branch.index + 1; k < c.Branch.index; k++ {
							c.Parent.graph[k].Connect.Set(api.BPass)
							c.Parent.graph[k].Branch.Set(api.BPass)
							if c.Parent.graph[k].PassColor == 0 {
								c.Parent.graph[k].PassColor = api.Color(b.color)
							} else {
								c.Parent.graph[k].PassColor = api.Color(cui.CWhite)
							}
						}
						// │
						for j := c.Index + 1; j < c.Parent.Index; j++ {
							cc := repo.Commits[j]
							cc.graph[i].Connect.Set(api.BMLine)
						}
						// ╯
						c.Parent.graph[i].Connect.Set(api.BBranchRight)
					} else {
						// Other branch is right side ╮
						c.graph[i+1].Connect.Set(api.BMergeRight)
						// │
						for j := c.Index + 1; j < c.Parent.Index; j++ {
							cc := repo.Commits[j]
							cc.graph[i+1].Connect.Set(api.BMLine)
						}
						//  ╰
						c.Parent.graph[c.Parent.Branch.index].Connect.Set(api.BBranchLeft)
						c.Parent.graph[c.Parent.Branch.index].Branch.Set(api.BBranchLeft)
					}
				}
			} else {
				// Other branch
				if b.tip == c {
					// this branch tip does not have a branch of its own, ┺
					c.graph[i].Branch.Set(api.BBottom | api.BPass)
					// ──
					for k := c.Branch.index + 1; k <= i; k++ {
						c.graph[k].Connect.Set(api.BPass)
						c.graph[k].Branch.Set(api.BPass)
						if c.graph[k].PassColor == 0 {
							c.graph[k].PassColor = api.Color(b.color)
						} else {
							c.graph[k].PassColor = api.Color(cui.CWhite)
						}
					}
				} else if c.Index >= b.tip.Index && c.Index <= b.bottom.Index {
					// Other branch, normal branch line (no commit on that branch) ┃
					c.graph[i].Branch.Set(api.BLine)
				}
			}
		}
	}
}
