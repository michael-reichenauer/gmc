package viewmodel

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
			c.graph[branch.index].Branch.Set(BTip) //       ┏   (branch tip)
		}
		if c == c.Branch.tip && c.Branch.isGitBranch {
			c.graph[branch.index].Branch.Set(BActiveTip) // ┣   (indicate possible more commits in the future)
		}
		if c == c.Branch.bottom {
			c.graph[branch.index].Branch.Set(BBottom) // ┗   (bottom commit (e.g. initial commit on master)
		}
		if c != c.Branch.tip && c != c.Branch.bottom { // ┣   (normal commit, in the middle)
			c.graph[branch.index].Branch.Set(BCommit)
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
		// if c.ID == StatusID {
		// 	continue
		// }
		for i, b := range repo.Branches {
			c.graph[i].BranchName = b.name
			c.graph[i].BranchDisplayName = b.displayName
			if c.Branch == b {
				// Commit branch
				if c.MergeParent != nil {
					// Commit has merge (2 parents)
					if c.MergeParent.Branch.index < c.Branch.index {
						// Other branch is left side ╭
						c.graph[i].Connect.Set(BMergeLeft)
						c.graph[i].Branch.Set(BMergeLeft)
						// Draw horizontal pass through line ───
						for k := c.MergeParent.Branch.index + 1; k < c.Branch.index; k++ {
							c.MergeParent.graph[k].Connect.Set(BPass)
							c.MergeParent.graph[k].Branch.Set(BPass)
							if c.MergeParent.graph[k].PassName == "" {
								c.MergeParent.graph[k].PassName = b.displayName
							} else {
								c.MergeParent.graph[k].PassName = "-"
							}
						}
						// Draw vertical down line │
						for j := c.Index + 1; j < c.MergeParent.Index; j++ {
							cc := repo.Commits[j]
							cc.graph[i].Connect.Set(BMLine)
						}
						// Draw ╯
						c.MergeParent.graph[i].Connect.Set(BBranchRight)
					} else {
						// Other branch is right side  ╮
						// Draw merge in rune
						c.graph[c.MergeParent.Branch.index].Connect.Set(BMergeRight)
						// Draw horizontal pass through line ───
						for k := i + 1; k < c.MergeParent.Branch.index; k++ {
							c.graph[k].Connect.Set(BPass)
							c.graph[k].Branch.Set(BPass)
							if c.graph[k].PassName == "" {
								c.graph[k].PassName = c.MergeParent.Branch.displayName
							} else {
								c.graph[k].PassName = "-"
							}
						}
						// Draw vertical down line │
						for j := c.Index + 1; j < c.MergeParent.Index; j++ {
							cc := repo.Commits[j]
							cc.graph[c.MergeParent.Branch.index].Connect.Set(BMLine)
						}
						// Draw branch out rune ╰
						c.MergeParent.graph[c.MergeParent.Branch.index].Connect.Set(BBranchLeft)
						c.MergeParent.graph[c.MergeParent.Branch.index].Branch.Set(BBranchLeft)
					}
				}
				if c.Parent != nil && c.Parent.Branch != c.Branch {
					// Commit parent is on other branch (bottom/first commit on this branch)
					if c.Parent.Branch.index < c.Branch.index {
						// Other branch is left side  ╭
						c.graph[i].Connect.Set(BMergeLeft)
						c.graph[i].Branch.Set(BMergeLeft)

						// ──
						for k := c.Parent.Branch.index + 1; k < c.Branch.index; k++ {
							c.Parent.graph[k].Connect.Set(BPass)
							c.Parent.graph[k].Branch.Set(BPass)
							if c.Parent.graph[k].PassName == "" {
								c.Parent.graph[k].PassName = b.displayName
							} else {
								c.Parent.graph[k].PassName = "-"
							}
						}
						// │
						for j := c.Index + 1; j < c.Parent.Index; j++ {
							cc := repo.Commits[j]
							cc.graph[i].Connect.Set(BMLine)
						}
						// ╯
						c.Parent.graph[i].Connect.Set(BBranchRight)
					} else {
						// Other branch is right side ╮
						c.graph[i+1].Connect.Set(BMergeRight)
						// │
						for j := c.Index + 1; j < c.Parent.Index; j++ {
							cc := repo.Commits[j]
							cc.graph[i+1].Connect.Set(BMLine)
						}
						//  ╰
						c.Parent.graph[c.Parent.Branch.index].Connect.Set(BBranchLeft)
						c.Parent.graph[c.Parent.Branch.index].Branch.Set(BBranchLeft)
					}
				}
			} else {
				// Other branch
				if b.tip == c {
					// this branch tip does not have a branch of its own, ┺
					c.graph[i].Branch.Set(BBottom | BPass)
					// ──
					for k := c.Branch.index + 1; k <= i; k++ {
						c.graph[k].Connect.Set(BPass)
						c.graph[k].Branch.Set(BPass)
						if c.graph[k].PassName == "" {
							c.graph[k].PassName = b.displayName
						} else {
							c.graph[k].PassName = "-"
						}
					}
				} else if c.Index >= b.tip.Index && c.Index <= b.bottom.Index {
					// Other branch, normal branch line (no commit on that branch) ┃
					c.graph[i].Branch.Set(BLine)
				}
			}
		}
	}
}
