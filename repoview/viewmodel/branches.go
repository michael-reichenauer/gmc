package viewmodel

import (
	"github.com/michael-reichenauer/gmc/repoview/viewmodel/gitrepo"
	"github.com/michael-reichenauer/gmc/utils"
	"sort"
)

func (h *Model) getGitModelBranches(branchNames []string, gmRepo gitrepo.Repo, gmStatus gitrepo.Status) []*gitrepo.Branch {
	if len(branchNames) == 0 {
		// No specified branches, default to current, or master
		branchNames = h.getDefaultBranchIDs(gmRepo)
	}

	var branches []*gitrepo.Branch
	for _, name := range branchNames {
		branch, ok := gmRepo.BranchByName(name)
		if ok {
			branches = append(branches, branch)
		}
	}

	branches = h.addLocalBranches(branches, gmRepo)
	branches = h.addRemoteBranches(branches, gmRepo)
	branches = h.removeSameLocalAsRemotes(branches, gmRepo, gmStatus)
	h.sortBranches(branches)
	return branches
}

func (h *Model) addLocalBranches(branches []*gitrepo.Branch, gmRepo gitrepo.Repo) []*gitrepo.Branch {
	var bs []*gitrepo.Branch
	for _, branch := range branches {
		bs = append(bs, branch)
		if branch.LocalName != "" {
			if !h.containsBranch(branches, branch.LocalName) {
				b, ok := gmRepo.BranchByName(branch.LocalName)
				if ok {
					bs = append(bs, b)
				}
			}
		}
	}
	return bs
}

func (h *Model) addRemoteBranches(branches []*gitrepo.Branch, gmRepo gitrepo.Repo) []*gitrepo.Branch {
	var bs []*gitrepo.Branch
	for _, branch := range branches {
		bs = append(bs, branch)
		if branch.RemoteName != "" {
			if !h.containsBranch(branches, branch.RemoteName) {
				b, ok := gmRepo.BranchByName(branch.RemoteName)
				if ok {
					bs = append(bs, b)
				}
			}
		}
	}
	return bs
}

func (h *Model) containsBranch(branches []*gitrepo.Branch, name string) bool {
	for _, b := range branches {
		if name == b.Name {
			return true
		}
	}
	return false
}

//func (h *Model) containsLocalBranch(branches []*gitrepo.Branch, name string) bool {
//	for _, b := range branches {
//		if name == b.RemoteName {
//			return true
//		}
//	}
//	return false
//}
func (h *Model) sortBranches(branches []*gitrepo.Branch) {
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

func (h *Model) getDefaultBranchIDs(gmRepo gitrepo.Repo) []string {
	var branchIDs []string
	branch, ok := gmRepo.CurrentBranch()
	if ok {
		return h.addBranchWithAncestors(branchIDs, branch)
	}
	branch, ok = gmRepo.BranchByName(remoteMasterName)
	if ok {
		return h.addBranchWithAncestors(branchIDs, branch)
	}
	branch, ok = gmRepo.BranchByName(masterName)
	if ok {
		return h.addBranchWithAncestors(branchIDs, branch)
	}
	return branchIDs
}

func (h *Model) addBranchWithAncestors(branchIds []string, branch *gitrepo.Branch) []string {
	ids := h.branchAncestorIDs(branch)
	for _, id := range ids {
		branchIds = h.addBranchIdIfNotAdded(branchIds, id)
	}
	return branchIds
}

func (*Model) addBranchIdIfNotAdded(branchIds []string, branchId string) []string {
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

func (h *Model) branchAncestorIDs(b *gitrepo.Branch) []string {
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

func (h *Model) removeSameLocalAsRemotes(branches []*gitrepo.Branch, gmRepo gitrepo.Repo, gmStatus gitrepo.Status) []*gitrepo.Branch {
	statusOk := gmStatus.OK()
	currentBranch, _ := gmRepo.CurrentBranch()

	var bs []*gitrepo.Branch
	for _, branch := range branches {
		if branch.RemoteName != "" &&
			h.containsSameRemoteBranch(branches, branch) &&
			!(!statusOk && branch == currentBranch) {
			continue
		}
		bs = append(bs, branch)
	}

	return bs
}

func (h *Model) containsSameRemoteBranch(bs []*gitrepo.Branch, branch *gitrepo.Branch) bool {
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
