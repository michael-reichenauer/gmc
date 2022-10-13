package viewrepo

import (
	"github.com/michael-reichenauer/gmc/api"

	//"github.com/michael-reichenauer/gmc/client/console"
	"github.com/michael-reichenauer/gmc/utils/cui"
	"github.com/michael-reichenauer/gmc/utils/log"
	"github.com/michael-reichenauer/gmc/utils/timer"
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
	s := timer.Start()
	t.drawBranchLines(repo)
	t.drawConnectorLines(repo)
	log.Infof("Graf %s", s)
}

func (s *branchesGraph) drawBranchLines(repo *repo) {
	for i, b := range repo.Branches {
		b.x = i
	}

	for _, b := range repo.Branches {
		c := b.tip

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

		for {
			x := c.Branch.x
			y := c.Index
			branchColor := b.color
			if c.Branch != b {
				// this commit is not part of the branch (several branches on the same commit)
				break
			}
			if c == c.Branch.tip {
				repo.SetGraphBranch(x, y, api.BTip, branchColor) //       ┏   (branch tip)
				//c.graph[b.index].Branch.Set(api.BTip) //       ┏   (branch tip)
			}
			if c == c.Branch.tip && c.Branch.isGitBranch {
				repo.SetGraphBranch(x, y, api.BActiveTip, branchColor) //
				//c.graph[b.index].Branch.Set(api.BActiveTip) // ┣   (indicate possible more commits in the future)
			}
			if c == c.Branch.bottom {
				repo.SetGraphBranch(x, y, api.BBottom, branchColor)
				//c.graph[b.index].Branch.Set(api.BBottom) //    ┗   (bottom commit (e.g. initial commit on main)
			}
			if c != c.Branch.tip && c != c.Branch.bottom { //       ┣   (normal commit, in the middle)
				repo.SetGraphBranch(x, y, api.BCommit, branchColor)
				//c.graph[b.index].Branch.Set(api.BCommit)
			}

			if c.Parent == nil || c.Branch != c.Parent.Branch {
				// Reached bottom of branch
				break
			}
			c = c.Parent
		}
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

func (s *branchesGraph) drawConnectorLines(repo *repo) {
	for _, c := range repo.Commits {
		for _, b := range repo.Branches {
			if c.Branch == b {
				// Commit branch
				if c.MergeParent != nil {
					s.drawMerge(repo, c)
				}
				if c.Parent != nil && c.Parent.Branch != c.Branch {
					// Commit parent is on other branch (i.e. commit is first/bottom commit on this branch)
					// Draw branched from parent branch
					s.drawBranchFromParent(repo, c)
				}
			} else {
				// Commit is on other branch
				x := b.x
				y := c.Index
				x2 := c.Branch.x
				color := b.color

				if b.tip == c {
					// this branch tip does not have a branch of its own, same row as parent, ┺
					repo.drawHorizontalLine(x2+1, x+1, y, color)           //              ─
					repo.SetGraphBranch(x, y, api.BBottom|api.Pass, color) //               ┺
				} else if c.Index >= b.tip.Index && c.Index <= b.bottom.Index {
					// Other branch, normal branch line (no commit on that branch)   ┃
					repo.SetGraphBranch(x, y, api.BLine, color) // ┃
				}
			}
		}
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

// for bi, b := range repo.Branches {
// 	if bi == 0 {
// 		continue
// 	}
// 	log.Infof("Branch %q", b.name)
// 	ct := repo.commitById[b.tipId]
// 	cb := repo.commitById[b.bottomId]
// 	firstIndex := ct.Index
// 	lastIndex := cb.Index
// 	childId, ok := lo.Find(ct.ChildIDs, func(v string) bool { return repo.commitById[v].MergeParent == ct })
// 	if ok {
// 		firstIndex = repo.commitById[childId].Index
// 	}
// 	if cb.Parent != nil {
// 		lastIndex = cb.Parent.Index
// 	}

// 	_, ok = lo.Find(repo.Commits[firstIndex:lastIndex], func(v *commit) bool {
// 		g := v.graph[bi-1]
// 		return !(g.Connect == 8 && g.Branch == 8 || g.Connect == 0 && g.Branch == 0)
// 	})
// 	if ok {
// 		continue
// 	}

// 	log.Infof("Branch %q can be moved", b.name)
// 	for i := firstIndex; i <= lastIndex; i++ {
// 		c := repo.Commits[i]
// 		graph := console.NewRepoGraph()
// 		var sb strings.Builder
// 		graph.WriteGraph(&sb, c.graph)
// 		cg := sb.String()
// 		g := c.graph

// 		log.Infof("%s %s %v, %s", cg, c.SID, g[bi-1], c.Subject)
// 	}

// 	for i := firstIndex; i <= lastIndex; i++ {
// 		c := repo.Commits[i]
// 		c.graph = append(c.graph[:bi-1], c.graph[bi:]...)
// 		c.graph = append(c.graph, api.GraphColumn{})
// 	}
// 	for i := 0; i < len(repo.Commits); i++ {
// 		c := repo.Commits[i]
// 		c.graph = append(c.graph[:bi], c.graph[bi:]...)
// 		c.graph = append(c.graph, api.GraphColumn{})
// 	}
// }

// for _, c := range repo.Commits {
// 	graph := console.NewRepoGraph()
// 	var sb strings.Builder
// 	graph.WriteGraph(&sb, c.graph)
// 	cg := sb.String()
// 	g := c.graph

// 	log.Infof("%s %s %v, %s", cg, c.SID, g, c.Subject)
// }
