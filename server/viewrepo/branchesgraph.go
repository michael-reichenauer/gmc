package viewrepo

import (
	"github.com/michael-reichenauer/gmc/api"

	//"github.com/michael-reichenauer/gmc/client/console"
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
	t.drawBranchLines(repo)
	t.drawConnectorLines(repo)
}

func (s *branchesGraph) drawBranchLines(repo *repo) {
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
			if c.Branch != b {
				// this commit is not part of the branch (several branches on the same commit)
				break
			}
			if c == c.Branch.tip {
				c.graph[b.index].Branch.Set(api.BTip) //       ┏   (branch tip)
			}
			if c == c.Branch.tip && c.Branch.isGitBranch {
				c.graph[b.index].Branch.Set(api.BActiveTip) // ┣   (indicate possible more commits in the future)
			}
			if c == c.Branch.bottom {
				c.graph[b.index].Branch.Set(api.BBottom) //    ┗   (bottom commit (e.g. initial commit on main)
			}
			if c != c.Branch.tip && c != c.Branch.bottom { //       ┣   (normal commit, in the middle)
				c.graph[b.index].Branch.Set(api.BCommit)
			}

			if c.Parent == nil || c.Branch != c.Parent.Branch {
				// Reached bottom of branch
				break
			}
			c = c.Parent
		}
	}
}

func (s *branchesGraph) drawConnectorLines(repo *repo) {
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
						if c.Branch != c.MergeParent.Branch {
							for j := c.Index + 1; j < c.MergeParent.Index; j++ {
								cc := repo.Commits[j]
								cc.graph[i].Connect.Set(api.BMLine)
							}
						}
						// Draw ╯
						c.MergeParent.graph[i].Connect.Set(api.BBranchRight)
					} else {
						// Other branch is right side  ╮
						// Draw merge in rune
						if c.Branch != c.MergeParent.Branch {
							c.graph[c.MergeParent.Branch.index].Connect.Set(api.BMergeRight)
						}
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
						if c.Branch != c.MergeParent.Branch {
							// Draw vertical down line │
							for j := c.Index + 1; j < c.MergeParent.Index; j++ {
								cc := repo.Commits[j]
								cc.graph[c.MergeParent.Branch.index].Connect.Set(api.BMLine)
							}
						}
						// Draw branch out rune ╰
						if c.Branch != c.MergeParent.Branch {
							c.MergeParent.graph[c.MergeParent.Branch.index].Connect.Set(api.BBranchLeft)
							c.MergeParent.graph[c.MergeParent.Branch.index].Branch.Set(api.BBranchLeft)
						} else {
							c.MergeParent.graph[c.MergeParent.Branch.index].Branch.Set(api.BCommit)
						}
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
}
