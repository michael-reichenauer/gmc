package viewrepo

import (
	"sort"

	"github.com/michael-reichenauer/gmc/server/viewrepo/augmented"
	"github.com/michael-reichenauer/gmc/utils"
	"github.com/michael-reichenauer/gmc/utils/git"
)

// getGitModelBranches returns the git branches based on the name together with ancestor branches
// If no named branches, the current branch (with ancestors is returned)
func (t *ViewRepoService) getAugmentedBranches(branchNames []string, augmentedRepo augmented.Repo) []*augmented.Branch {
	branchNames = t.getBranchNamesToShow(branchNames, augmentedRepo)

	var branches []*augmented.Branch
	var notFound []string
	for _, name := range branchNames {
		branch, ok := augmentedRepo.BranchByName(name)
		if ok {
			for _, b := range branch.GetAncestorsAndSelf() {
				branches = t.appendIfNotAppended(branches, b)
			}
		} else {
			notFound = append(notFound, name)
		}
	}
	// Check if some names may have been deleted branches, which now have different names,
	// but same display name
	for _, name := range notFound {
		n := git.StripRemotePrefix(name)
		for _, b := range augmentedRepo.Branches {
			if n == b.DisplayName {
				for _, b := range b.GetAncestorsAndSelf() {
					branches = t.appendIfNotAppended(branches, b)
				}
				break
			}
		}
	}

	branches = t.addLocalBranches(branches, augmentedRepo)
	branches = t.addRemoteBranches(branches, augmentedRepo)
	t.sortBranches(branches)
	return branches
}

func (t *ViewRepoService) setBranchColors(repo *repo) {
	for _, b := range repo.Branches {
		b.color = t.BranchColor(b.displayName)
	}
}

func (t *ViewRepoService) appendIfNotAppended(branches []*augmented.Branch, branch *augmented.Branch) []*augmented.Branch {
	for _, b := range branches {
		if b == branch {
			return branches
		}
	}
	return append(branches, branch)
}

func (t *ViewRepoService) addLocalBranches(branches []*augmented.Branch, gmRepo augmented.Repo) []*augmented.Branch {
	var bs []*augmented.Branch
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

func (t *ViewRepoService) getBranchNamesToShow(branchNames []string, augmentedRepo augmented.Repo) []string {
	if len(branchNames) > 0 {
		return branchNames
	}

	// No specified branches, default to current, or main
	rc := t.configService.GetRepo(t.augmentedRepo.RepoPath())
	branchNames = rc.ShownBranches
	if len(branchNames) == 0 {
		branchNames = t.getDefaultBranchIDs(augmentedRepo)
	}

	return branchNames
}

func (t *ViewRepoService) addRemoteBranches(branches []*augmented.Branch, gmRepo augmented.Repo) []*augmented.Branch {
	var bs []*augmented.Branch
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

func (t *ViewRepoService) containsBranch(branches []*augmented.Branch, name string) bool {
	for _, b := range branches {
		if name == b.Name {
			return true
		}
	}
	return false
}

func (t *ViewRepoService) sortBranches(branches []*augmented.Branch) {
	sort.SliceStable(branches, func(l, r int) bool {
		left := branches[l]
		right := branches[r]

		if left.Name == right.RemoteName {
			// left branch is the remote branch of the right branch
			// Prioritize remote branch before local
			return true
		}

		// Prioritize known branches like master, develop
		il := utils.StringsIndex(augmented.DefaultBranchPriority, left.Name)
		ir := utils.StringsIndex(augmented.DefaultBranchPriority, right.Name)
		if il != -1 && (il < ir || ir == -1) {
			// Left item is known branch with higher priority
			return true
		}

		if right.IsAncestorBranch(left.Name) {
			return true
		}

		// no known order for the pair
		return false
	})
}

func (t *ViewRepoService) getDefaultBranchIDs(gmRepo augmented.Repo) []string {
	var branchIDs []string
	branch, ok := gmRepo.CurrentBranch()
	if ok {
		return t.addBranchWithAncestors(branchIDs, branch)
	}
	branch, ok = gmRepo.BranchByName(remoteMainName)
	if ok {
		return t.addBranchWithAncestors(branchIDs, branch)
	}
	branch, ok = gmRepo.BranchByName(mainName)
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

func (t *ViewRepoService) addBranchWithAncestors(branchIds []string, branch *augmented.Branch) []string {
	ids := t.branchAncestorIDs(branch)
	for _, id := range ids {
		branchIds = t.addBranchIdIfNotAdded(branchIds, id)
	}
	return branchIds
}

func (*ViewRepoService) addBranchIdIfNotAdded(branchIds []string, branchId string) []string {
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

func (*ViewRepoService) branchAncestorIDs(b *augmented.Branch) []string {
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

func (t *ViewRepoService) removeSameLocalAsRemotes(branches []*augmented.Branch, gmRepo augmented.Repo, gmStatus augmented.Status) []*augmented.Branch {
	statusOk := gmStatus.OK()
	currentBranch, _ := gmRepo.CurrentBranch()

	var bs []*augmented.Branch
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

func (*ViewRepoService) containsSameRemoteBranch(bs []*augmented.Branch, branch *augmented.Branch) bool {
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
