package model

import (
	"github.com/michael-reichenauer/gmc/repoview/model/gitmodel"
	"github.com/michael-reichenauer/gmc/utils"
	"sort"
)

func (h *Model) getGitModelBranches(branchIds []string, gmRepo gitmodel.Repo, gmStatus gitmodel.Status) []*gitmodel.Branch {
	if len(branchIds) == 0 {
		// No specified branches, default to current, or master
		branchIds = h.getDefaultBranchIDs(gmRepo)
	}

	var branches []*gitmodel.Branch
	for _, id := range branchIds {
		branch, ok := gmRepo.BranchByID(id)
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

func (h *Model) addLocalBranches(branches []*gitmodel.Branch, gmRepo gitmodel.Repo) []*gitmodel.Branch {
	var bs []*gitmodel.Branch
	for _, branch := range branches {
		bs = append(bs, branch)
		if branch.IsRemote {
			if !h.containsLocalBranch(branches, branch.Name) {
				b, ok := gmRepo.LocalBranchByRemoteName(branch.Name)
				if ok {
					bs = append(bs, b)
				}
			}
		}
	}
	return bs
}

func (h *Model) addRemoteBranches(branches []*gitmodel.Branch, gmRepo gitmodel.Repo) []*gitmodel.Branch {
	var bs []*gitmodel.Branch
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
func (h *Model) containsBranch(branches []*gitmodel.Branch, name string) bool {
	for _, b := range branches {
		if name == b.Name {
			return true
		}
	}
	return false
}

func (h *Model) containsLocalBranch(branches []*gitmodel.Branch, name string) bool {
	for _, b := range branches {
		if name == b.RemoteName {
			return true
		}
	}
	return false
}
func (h *Model) sortBranches(branches []*gitmodel.Branch) {
	sort.SliceStable(branches, func(l, r int) bool {
		if branches[l].Name == branches[r].RemoteName {
			// Prioritize remote branch before local
			return true
		}
		// Prioritize known branches like master, develop
		il := utils.StringsIndex(gitmodel.DefaultBranchPrio, branches[l].Name)
		ir := utils.StringsIndex(gitmodel.DefaultBranchPrio, branches[r].Name)
		if il != -1 && (il < ir || ir == -1) {
			// Left item is known branch with higher priority
			return true
		}
		// no known order for the pair
		return false
	})
}

func (h *Model) getDefaultBranchIDs(gmRepo gitmodel.Repo) []string {
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

func (h *Model) addBranchWithAncestors(branchIds []string, branch *gitmodel.Branch) []string {
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

func (h *Model) branchAncestorIDs(b *gitmodel.Branch) []string {
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

func (h *Model) removeSameLocalAsRemotes(branches []*gitmodel.Branch, gmRepo gitmodel.Repo, gmStatus gitmodel.Status) []*gitmodel.Branch {
	statusOk := gmStatus.OK()
	currentBranch, _ := gmRepo.CurrentBranch()

	var bs []*gitmodel.Branch
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

func (h *Model) containsSameRemoteBranch(bs []*gitmodel.Branch, branch *gitmodel.Branch) bool {
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
