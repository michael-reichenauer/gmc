package console

import (
	"fmt"

	"github.com/michael-reichenauer/gmc/api"
	"github.com/michael-reichenauer/gmc/utils/cui"
	"github.com/michael-reichenauer/gmc/utils/git"
	"github.com/michael-reichenauer/gmc/utils/log"
	"github.com/samber/lo"
	"github.com/thoas/go-funk"
)

type menuService struct {
	ui cui.UI
	vm *repoVM
}

func newMenuService(ui cui.UI, vm *repoVM) *menuService {
	return &menuService{ui: ui, vm: vm}
}

func (t *menuService) getContextMenu(currentLineIndex int) cui.Menu {
	log.Infof(">")
	defer log.Infof("<")
	menu := t.ui.NewMenu("")
	c := t.vm.repo.Commits[currentLineIndex]
	b := t.vm.repo.Branches[c.BranchIndex]

	menu.Add(cui.MenuSeparator(fmt.Sprintf("Commit: %s", c.SID)))
	menu.Add(cui.MenuItem{Text: "Show Details ...", Action: func() { t.vm.showCommitDetails() }})
	menu.Add(cui.MenuItem{Text: "Show Diff ...", Key: "D", Action: func() {
		t.vm.showCommitDiff(c.ID)
	}})
	menu.Add(cui.MenuItem{Text: "Commit ...", Key: "C", Action: t.vm.showCommitDialog})
	menu.Add(cui.MenuItem{Text: "Files Diffs", Title: "All Files", SubItemsFunc: func() []cui.MenuItem {
		return t.getFileDiffsMenuItems()
	}})

	menu.Add(cui.MenuSeparator("Branches"))
	menu.Add(cui.MenuItem{Text: "Show Branch", SubItemsFunc: func() []cui.MenuItem {
		return t.getShowBranchesMenuItems(currentLineIndex)
	}})

	menu.Add(cui.MenuItem{Text: "Hide Branch", SubItemsFunc: func() []cui.MenuItem {
		return t.getHideBranchMenuItems()
	}})
	menu.Add(cui.MenuItem{Text: "Push", Title: "Push", SubItemsFunc: func() []cui.MenuItem {
		return t.getPushBranchMenuItems()
	}})
	menu.Add(cui.MenuItem{Text: "Pull/Update", Title: "Update", SubItemsFunc: func() []cui.MenuItem {
		return t.getPullBranchMenuItems()
	}})
	menu.Add(cui.MenuItem{Text: "Create Branch ...", Key: "B", Action: t.vm.showCreateBranchDialog})
	menu.Add(cui.MenuItem{Text: "Delete Branch", SubItemsFunc: func() []cui.MenuItem {
		return t.getDeleteBranchMenuItems()
	}})
	menu.Add(cui.MenuItem{Text: "Switch/Checkout", Title: "Switch To", Key: "S", SubItemsFunc: func() []cui.MenuItem {
		return t.getSwitchBranchMenuItems()
	}})
	menu.Add(cui.MenuItem{Text: "Merge", Title: fmt.Sprintf("Merge Into: %s", t.vm.repo.CurrentBranchName), Key: "", SubItemsFunc: func() []cui.MenuItem {
		return t.getMergeMenuItems()
	}})

	menu.Add(cui.MenuItem{Text: "Search ...", Key: "F", Action: t.vm.ShowSearchView})

	// hierarchy
	if b.IsAmbiguousBranch || b.IsSetAsParent {
		mi := []cui.MenuItem{}
		if b.IsSetAsParent {
			mi = append(mi, cui.MenuItem{Text: "Unset Branch as Parent", Action: func() {
				t.vm.UnsetAsParentBranch(b.Name)
			}})
		}
		if b.IsAmbiguousBranch {
			mi = append(mi, cui.MenuItem{Text: "Set Ambiguous Branch Parent", SubItemsFunc: func() []cui.MenuItem {
				return t.getAmbiguousBranchBranchesMenuItems()
			}})
		}

		menu.Add(cui.MenuItem{Text: "Branch Hierarchy", SubItems: mi})
	}

	menu.Add(cui.MenuSeparator("Repos"))

	menu.Add(cui.MenuItem{Text: "Open Repo", Title: "Open", SubItemsFunc: func() []cui.MenuItem {
		return t.vm.repoViewer.OpenRepoMenuItems()
	}})

	// menu.Add(t.vm.mainService.MainMenuItem())
	return menu
}

func (t *menuService) getSwitchMenu() cui.Menu {
	menu := t.ui.NewMenu("Switch/Checkout")
	menu.AddItems(t.getSwitchBranchMenuItems())
	return menu
}

func (t *menuService) getMergeMenu(name string) cui.Menu {
	menu := t.ui.NewMenu(fmt.Sprintf("Merge Into: %s", name))
	menu.AddItems(t.getMergeMenuItems())
	return menu
}

func (t *menuService) getShowHideBranchesMenu() cui.Menu {
	menu := t.ui.NewMenu("Hide Branch")
	menu.AddItems(t.getHideBranchMenuItems())
	return menu
}

func (t *menuService) getShowCommitBranchesMenu(selectedIndex int) cui.Menu {
	menu := t.ui.NewMenu("Show branch")
	menu.AddItems(t.getShowCommitBranchesMenuItems(selectedIndex))
	// menu.Add(cui.MenuItem{Text: "Hide Branch", SubItems: t.getHideBranchMenuItems()})
	return menu
}

func (t *menuService) getShowCommitBranchesMenuItems(selectedIndex int) []cui.MenuItem {
	var items []cui.MenuItem
	branches := t.vm.GetCommitBranches(selectedIndex)
	for _, b := range branches {
		items = append(items, t.toOpenBranchMenuItem(b))
	}
	return items
}

func (t *menuService) getShowBranchesMenuItems(selectedIndex int) []cui.MenuItem {
	ambiguousBranches := t.vm.GetNotShownAmbiguousBranches()
	branches := t.vm.GetCommitBranches(selectedIndex)
	var items []cui.MenuItem
	current, ok := t.vm.CurrentNotShownBranch()
	if ok {
		if nil == funk.Find(branches, func(b api.Branch) bool {
			return current.DisplayName == b.DisplayName
		}) {
			items = append(items, t.toOpenBranchMenuItem(current))
		}
	}

	if len(items) > 0 {
		items = append(items, cui.MenuSeparator(""))
	}

	items = append(items, cui.MenuItem{Text: "Latest Branches", SubItemsFunc: func() []cui.MenuItem {
		var latestSubItems []cui.MenuItem
		for _, b := range t.vm.GetLatestBranches(true) {
			latestSubItems = append(latestSubItems, t.toOpenBranchMenuItem(b))
		}
		return latestSubItems
	}})

	items = append(items, cui.MenuItem{Text: "Live Branches", SubItemsFunc: func() []cui.MenuItem {
		var allGitSubItems []cui.MenuItem
		for _, b := range t.vm.GetAllBranches(true) {
			if b.IsGitBranch {
				allGitSubItems = append(allGitSubItems, t.toOpenBranchMenuItem(b))
			}
		}
		return allGitSubItems
	}})

	items = append(items, cui.MenuItem{Text: "Live and Deleted Branches", SubItemsFunc: func() []cui.MenuItem {
		var allSubItems []cui.MenuItem
		for _, b := range t.vm.GetAllBranches(true) {
			allSubItems = append(allSubItems, t.toOpenBranchMenuItem(b))
		}
		return allSubItems
	}})

	if len(ambiguousBranches) > 0 {
		items = append(items, cui.MenuItem{Text: "Ambiguous Branches", SubItemsFunc: func() []cui.MenuItem {
			var allSubItems []cui.MenuItem
			for _, b := range t.vm.GetNotShownAmbiguousBranches() {
				allSubItems = append(allSubItems, t.toOpenBranchMenuItem(b))
			}
			return allSubItems
		}})
	}

	return items
}

func (t *menuService) getHideBranchMenuItems() []cui.MenuItem {
	var items []cui.MenuItem
	commitBranches := t.vm.GetShownBranches(true)
	for _, b := range commitBranches {
		name := b.Name // closure save
		closeItem := cui.MenuItem{Text: t.branchItemText(b), Action: func() {
			t.vm.HideBranch(name)
		}}
		items = append(items, closeItem)
	}
	return items
}

func (t *menuService) getSwitchBranchMenuItems() []cui.MenuItem {
	var items []cui.MenuItem
	commitBranches := t.vm.GetShownBranches(false)
	for _, b := range commitBranches {
		if b.IsCurrent {
			continue
		}

		bb := b // closure save
		switchItem := cui.MenuItem{Text: t.branchItemText(b), Action: func() {
			t.vm.SwitchToBranch(bb.Name, bb.DisplayName)
		}}
		items = append(items, switchItem)
	}

	items = append(items, cui.MenuSeparator(""))

	items = append(items, cui.MenuItem{Text: "Latest Branches", SubItemsFunc: func() []cui.MenuItem {
		var activeSubItems []cui.MenuItem
		for _, b := range t.vm.GetLatestBranches(true) {
			bb := b // closure save
			switchItem := cui.MenuItem{Text: t.branchItemText(b), Action: func() {
				t.vm.SwitchToBranch(bb.Name, bb.DisplayName)
			}}
			activeSubItems = append(activeSubItems, switchItem)
		}
		return activeSubItems
	}})

	items = append(items, cui.MenuItem{Text: "Live Branches", SubItemsFunc: func() []cui.MenuItem {
		var allGitSubItems []cui.MenuItem
		for _, b := range t.vm.GetAllBranches(true) {
			bb := b // closure save
			if !b.IsGitBranch {
				continue
			}
			switchItem := cui.MenuItem{Text: t.branchItemText(b), Action: func() {
				t.vm.SwitchToBranch(bb.Name, bb.DisplayName)
			}}
			allGitSubItems = append(allGitSubItems, switchItem)
		}
		return allGitSubItems
	}})

	items = append(items, cui.MenuItem{Text: "Live and Deleted Branches", SubItemsFunc: func() []cui.MenuItem {
		var allSubItems []cui.MenuItem
		for _, b := range t.vm.GetAllBranches(true) {
			bb := b // closure save
			switchItem := cui.MenuItem{Text: t.branchItemText(b), Action: func() {
				t.vm.SwitchToBranch(bb.Name, bb.DisplayName)
			}}
			allSubItems = append(allSubItems, switchItem)
		}
		return allSubItems
	}})

	return items
}

func (t *menuService) toOpenBranchMenuItem(branch api.Branch) cui.MenuItem {
	text := t.branchItemText(branch)
	if !branch.IsGitBranch {
		// Not a git branch, mark the branch a bit darker
		text = cui.Dark(text)
	}

	return cui.MenuItem{Text: text, Action: func() {
		t.vm.ShowBranch(branch.Name)
	}}
}

func (t *menuService) toSetAsParentBranchMenuItem(branch api.Branch) cui.MenuItem {
	text := t.branchItemText(branch)

	return cui.MenuItem{Text: text, Action: func() {
		t.vm.SetAsParentBranch(branch.Name)
	}}
}

func (t *menuService) branchItemText(branch api.Branch) string {
	prefix := " "
	if branch.IsIn {
		prefix = "╮"
	} else if branch.IsOut {
		prefix = "╭"
	}
	if branch.IsCurrent {
		return prefix + "●" + branch.DisplayName
	} else {
		return prefix + " " + branch.DisplayName
	}
}

func (t *menuService) getPushBranchMenuItems() []cui.MenuItem {
	var items []cui.MenuItem
	current, ok := t.vm.CurrentBranch()
	if ok && current.HasLocalOnly {
		pushItem := cui.MenuItem{Text: t.branchItemText(current), Key: "P", Action: func() {
			t.vm.PushBranch(current.Name)
		}}
		items = append(items, pushItem)
	}

	// Add other branches if they have commits (but only if the can be pushed cleanly)
	var otherItems []cui.MenuItem
	for _, b := range t.vm.repo.Branches {
		if !b.IsCurrent && !b.IsRemote && b.HasLocalOnly && !b.HasRemoteOnly {
			bClosure := b
			pushItem := cui.MenuItem{Text: t.branchItemText(bClosure), Action: func() {
				t.vm.PushBranch(bClosure.DisplayName)
			}}
			otherItems = append(otherItems, pushItem)
		}
	}

	// Add separator between current and other branches
	if len(items) > 0 && len(otherItems) > 0 {
		items = append(items, cui.MenuSeparator(""))
	}

	items = append(items, otherItems...)

	return items
}

func (t *menuService) getPullBranchMenuItems() []cui.MenuItem {
	var items []cui.MenuItem
	current, ok := t.vm.CurrentBranch()

	// Add current branch if it has commits that can be pulled
	if ok && current.HasRemoteOnly {
		pushItem := cui.MenuItem{Text: t.branchItemText(current), Key: "U", Action: func() {
			t.vm.PullCurrentBranch()
		}}
		items = append(items, pushItem)
	}

	// Add other branches if they have commits (but only if the can be pulled cleanly)
	var otherItems []cui.MenuItem
	for _, b := range t.vm.repo.Branches {
		if !b.IsCurrent && b.IsRemote && b.HasRemoteOnly && !b.HasLocalOnly {
			bClosure := b
			pushItem := cui.MenuItem{Text: t.branchItemText(bClosure), Action: func() {
				t.vm.PullBranch(bClosure.DisplayName)
			}}
			otherItems = append(otherItems, pushItem)
		}
	}

	// Add separator between current and other branches
	if len(items) > 0 && len(otherItems) > 0 {
		items = append(items, cui.MenuSeparator(""))
	}

	items = append(items, otherItems...)
	return items
}

func (t *menuService) getMergeMenuItems() []cui.MenuItem {
	current, ok := t.vm.CurrentBranch()
	if !ok {
		return nil
	}
	var items []cui.MenuItem
	commitBranches := t.vm.GetShownBranches(false)
	for _, b := range commitBranches {
		name := b.Name // closure save
		if b.DisplayName == current.DisplayName {
			continue
		}
		item := cui.MenuItem{Text: t.branchItemText(b), Action: func() {
			t.vm.MergeFromBranch(name)
		}}
		items = append(items, item)
	}
	return items
}

func (t *menuService) getFileDiffsMenuItems() []cui.MenuItem {
	c := t.vm.repo.Commits[t.vm.currentIndex]
	ref := c.ID
	if c.ID == git.UncommittedID {
		cb, ok := t.vm.CurrentBranch()
		if !ok {
			return []cui.MenuItem{}
		}
		ref = cb.Name
	}

	files := t.vm.GetFiles(ref)
	return lo.Map(files, func(v string, _ int) cui.MenuItem {
		return cui.MenuItem{Text: v, Action: func() {
			t.vm.showFileDiff(v)
		}}
	})
}

func (t *menuService) getDeleteBranchMenuItems() []cui.MenuItem {
	var items []cui.MenuItem
	branches := t.vm.GetAllBranches(false)
	for _, b := range branches {
		if !b.IsGitBranch || b.IsMainBranch || b.IsCurrent {
			// Do not delete main branch
			continue
		}
		name := b.Name // closure save
		item := cui.MenuItem{Text: t.branchItemText(b), Action: func() {
			t.vm.DeleteBranch(name)
		}}
		items = append(items, item)
	}
	return items

}

func (t *menuService) getAmbiguousBranchBranchesMenuItems() []cui.MenuItem {
	var items []cui.MenuItem

	for _, b := range t.vm.GetAmbiguousBranchBranchesMenuItems() {
		items = append(items, t.toSetAsParentBranchMenuItem(b))
	}

	return items
}
