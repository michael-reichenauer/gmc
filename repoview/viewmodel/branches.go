package viewmodel

import (
	"github.com/michael-reichenauer/gmc/repoview/viewmodel/gitrepo"
	"github.com/michael-reichenauer/gmc/utils"
	"sort"
)

// getGitModelBranches returns the git branches based on the name together with ancestor branches
// If no named branches, the current branch (with ancestors is returned)
func (t *Service) getGitRepoBranches(branchNames []string, gmRepo gitrepo.Repo) []*gitrepo.Branch {
	if len(branchNames) == 0 {
		// No specified branches, default to current, or master
		rc := t.configService.GetRepo(t.gitRepo.RepoPath())
		branchNames = rc.ShownBranches
		if len(branchNames) == 0 {
			branchNames = t.getDefaultBranchIDs(gmRepo)
		}
	}

	var branches []*gitrepo.Branch
	for _, name := range branchNames {
		branch, ok := gmRepo.BranchByName(name)
		if ok {
			branches = append(branches, branch)
		}
	}

	branches = t.addLocalBranches(branches, gmRepo)
	branches = t.addRemoteBranches(branches, gmRepo)
	// branches = s.removeSameLocalAsRemotes(branches, gmRepo, gmStatus)
	t.sortBranches(branches)
	return branches
}

func (t *Service) addLocalBranches(branches []*gitrepo.Branch, gmRepo gitrepo.Repo) []*gitrepo.Branch {
	var bs []*gitrepo.Branch
	for _, branch := range branches {
		bs = append(bs, branch)
		if branch.LocalName != "" {
			if !t.containsBranch(branches, branch.LocalName) {
				b, ok := gmRepo.BranchByName(branch.LocalName)
				if ok {
					bs = append(bs, b)
				}
			}
		}
	}
	return bs
}

func (t *Service) addRemoteBranches(branches []*gitrepo.Branch, gmRepo gitrepo.Repo) []*gitrepo.Branch {
	var bs []*gitrepo.Branch
	for _, branch := range branches {
		bs = append(bs, branch)
		if branch.RemoteName != "" {
			if !t.containsBranch(branches, branch.RemoteName) {
				b, ok := gmRepo.BranchByName(branch.RemoteName)
				if ok {
					bs = append(bs, b)
				}
			}
		}
	}
	return bs
}

func (t *Service) containsBranch(branches []*gitrepo.Branch, name string) bool {
	for _, b := range branches {
		if name == b.Name {
			return true
		}
	}
	return false
}

func (t *Service) sortBranches(branches []*gitrepo.Branch) {
	sort.SliceStable(branches, func(l, r int) bool {
		if branches[l].Name == branches[r].RemoteName {
			// Prioritize remote branch before local
			return true
		}
		// Prioritize known branches like master, develop
		il := utils.StringsIndex(gitrepo.DefaultBranchPriority, branches[l].Name)
		ir := utils.StringsIndex(gitrepo.DefaultBranchPriority, branches[r].Name)
		if il != -1 && (il < ir || ir == -1) {
			// Left item is known branch with higher priority
			return true
		}
		// no known order for the pair
		return false
	})
}

func (t *Service) getDefaultBranchIDs(gmRepo gitrepo.Repo) []string {
	var branchIDs []string
	branch, ok := gmRepo.CurrentBranch()
	if ok {
		return t.addBranchWithAncestors(branchIDs, branch)
	}
	branch, ok = gmRepo.BranchByName(remoteMasterName)
	if ok {
		return t.addBranchWithAncestors(branchIDs, branch)
	}
	branch, ok = gmRepo.BranchByName(masterName)
	if ok {
		return t.addBranchWithAncestors(branchIDs, branch)
	}
	return branchIDs
}

func (t *Service) addBranchWithAncestors(branchIds []string, branch *gitrepo.Branch) []string {
	ids := t.branchAncestorIDs(branch)
	for _, id := range ids {
		branchIds = t.addBranchIdIfNotAdded(branchIds, id)
	}
	return branchIds
}

func (*Service) addBranchIdIfNotAdded(branchIds []string, branchId string) []string {
	isAdded := false
	for _, id := range branchIds {
		if id == branchId {
			isAdded = true
		}
	}
	if !isAdded {
		branchIds = append(branchIds, branchId)
	}
	return branchIds
}

func (*Service) branchAncestorIDs(b *gitrepo.Branch) []string {
	var ids []string
	for cb := b; cb != nil; cb = cb.ParentBranch {
		ids = append(ids, cb.Name)
	}
	for i := len(ids)/2 - 1; i >= 0; i-- {
		opp := len(ids) - 1 - i
		ids[i], ids[opp] = ids[opp], ids[i]
	}
	return ids
}

func (t *Service) removeSameLocalAsRemotes(branches []*gitrepo.Branch, gmRepo gitrepo.Repo, gmStatus gitrepo.Status) []*gitrepo.Branch {
	statusOk := gmStatus.OK()
	currentBranch, _ := gmRepo.CurrentBranch()

	var bs []*gitrepo.Branch
	for _, branch := range branches {
		if branch.RemoteName != "" &&
			t.containsSameRemoteBranch(branches, branch) &&
			!(!statusOk && branch == currentBranch) {
			continue
		}
		bs = append(bs, branch)
	}

	return bs
}

func (*Service) containsSameRemoteBranch(bs []*gitrepo.Branch, branch *gitrepo.Branch) bool {
	for _, b := range bs {
		if branch.RemoteName != "" &&
			branch.RemoteName == b.Name &&
			branch.TipID == b.TipID {
			// branch is a local branch with same branch tip as the remote branch in the bs list
			return true
		}
	}
	return false
}
