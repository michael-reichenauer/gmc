package console

import (
	"fmt"

	"github.com/michael-reichenauer/gmc/api"
	"github.com/michael-reichenauer/gmc/utils/cui"
	"github.com/michael-reichenauer/gmc/utils/git"
	"github.com/samber/lo"
)

type Menus interface {
	GetContextMenu(currentLineIndex int) cui.Menu
	GetShowBranchesMenu(selectedIndex int) cui.Menu
	GetHideBranchesMenu() cui.Menu
	GetMergeMenu(name string) cui.Menu
}

type menus struct {
	ui cui.UI
	vm *repoVM
}

func newMenus(ui cui.UI, vm *repoVM) Menus {
	return &menus{ui: ui, vm: vm}
}

func (t *menus) GetContextMenu(currentLineIndex int) cui.Menu {
	menu := t.ui.NewMenu("Main Menu")
	menu.AddItems(t.getContextMenuItems(currentLineIndex))
	return menu
}

func (t *menus) GetShowBranchesMenu(selectedIndex int) cui.Menu {
	menu := t.ui.NewMenu("Branches")
	menu.Add(cui.MenuSeparator("Show"))
	menu.AddItems(t.getShowBranchesMenuItems(selectedIndex))
	menu.Add(cui.MenuSeparator("Switch to"))
	menu.AddItems(t.getSwitchBranchMenuItems(true))
	menu.Add(cui.MenuSeparator(""))
	menu.Add(cui.MenuItem{Text: "Main Menu", Title: "Main Menu", Key: "M", Items: t.getContextMenuItems(selectedIndex)})
	return menu
}

func (t *menus) GetHideBranchesMenu() cui.Menu {
	menu := t.ui.NewMenu("Hide Branch")
	menu.AddItems(t.getHideBranchMenuItems())
	return menu
}

func (t *menus) GetMergeMenu(name string) cui.Menu {
	menu := t.ui.NewMenu(fmt.Sprintf("Merge Into: %s", name))
	menu.AddItems(t.getMergeMenuItems())
	return menu
}

func (t *menus) showAbout() {
	t.ui.ShowMessageBox("About gmc", fmt.Sprintf("Version: %s", t.ui.Version()))
}

func (t *menus) getContextMenuItems(currentLineIndex int) []cui.MenuItem {
	c := t.vm.repo.Commits[currentLineIndex]
	items := []cui.MenuItem{}

	// Commit items
	items = append(items, cui.MenuSeparator(fmt.Sprintf("Commit: %s", c.SID)))
	items = append(items, cui.MenuItem{Text: "Toggle Details ...", Key: "Enter", Action: t.vm.repoViewer.ShowCommitDetails})
	items = append(items, cui.MenuItem{Text: "Commit ...", Key: "C", Action: t.vm.showCommitDialog})
	items = append(items, cui.MenuItem{Text: "Commit Diff ...", Key: "D", Action: func() { t.vm.showCommitDiff(c.ID) }})

	// Branches items
	items = append(items, cui.MenuSeparator("Branches"))
	items = append(items, cui.MenuItem{Text: "Show Branch", Title: "Show Branch", Key: "->", ItemsFunc: func() []cui.MenuItem {
		return t.getShowBranchesMenuItems(currentLineIndex)
	}})
	items = append(items, cui.MenuItem{Text: "Hide Branch", Title: "Hide Branch", Key: "<-", ItemsFunc: t.getHideBranchMenuItems})
	items = append(items, cui.MenuItem{Text: "Switch/Checkout", Title: "Switch To", ItemsFunc: func() []cui.MenuItem {
		return t.getSwitchBranchMenuItems(false)
	}})
	items = append(items, cui.MenuItem{Text: "Push", Title: "Push", ItemsFunc: t.getPushBranchMenuItems})
	items = append(items, cui.MenuItem{Text: "Pull/Update", Title: "Update", ItemsFunc: t.getPullBranchMenuItems})
	items = append(items, cui.MenuItem{Text: "Merge", Title: fmt.Sprintf("Merge Into: %s", t.vm.repo.CurrentBranchName), Key: "M",
		ItemsFunc: t.getMergeMenuItems})
	items = append(items, cui.MenuItem{Text: "Create Branch ...", Key: "B", Action: t.vm.showCreateBranchDialog})
	items = append(items, cui.MenuItem{Text: "Delete Branch", ItemsFunc: t.getDeleteBranchMenuItems})

	hi := t.getBranchHierarchyMenuItems(c)
	if len(items) > 0 {
		items = append(items, cui.MenuItem{Text: "Branch Hierarchy", Items: hi})
	}

	// Other items
	items = append(items, cui.MenuSeparator("More"))
	items = append(items, cui.MenuItem{Text: "Search/Filter ...", Key: "F", Action: t.vm.ShowSearchView})
	items = append(items, cui.MenuItem{Text: "File History", Title: "All Files", ItemsFunc: t.getFileDiffsMenuItems})
	items = append(items, cui.MenuItem{Text: "Open Repo", Title: "Open", ItemsFunc: t.vm.repoViewer.OpenRepoMenuItems})
	items = append(items, cui.MenuItem{Text: "About ...", Action: t.showAbout})

	return items
}

func (t *menus) getShowBranchesMenuItems(selectedIndex int) []cui.MenuItem {
	ambiguousBranches := t.vm.GetNotShownAmbiguousBranches()
	branches := t.vm.GetCommitBranches(selectedIndex)
	items := []cui.MenuItem{}

	current, ok := t.vm.CurrentNotShownBranch()
	if ok {
		// Current branch is not already shown
		_, ok = lo.Find(branches, func(b api.Branch) bool {
			return current.DisplayName == b.DisplayName
		})
		if !ok {
			// Current branch is not amongst the commit branches
			items = append(items, t.toOpenBranchMenuItem(current))
		}
	}

	items = append(items, t.getShowCommitBranchesMenuItems(selectedIndex)...)

	if len(items) > 0 {
		items = append(items, cui.MenuSeparator(""))
	}

	items = append(items, cui.MenuItem{Text: "Recent Branches", ItemsFunc: func() []cui.MenuItem {
		return lo.Map(t.vm.GetRecentBranches(true), func(v api.Branch, _ int) cui.MenuItem {
			return t.toOpenBranchMenuItem(v)
		})
	}})

	items = append(items, cui.MenuItem{Text: "Live Branches", ItemsFunc: func() []cui.MenuItem {
		var allGitSubItems []cui.MenuItem
		for _, b := range t.vm.GetAllBranches(true) {
			if b.IsGitBranch {
				allGitSubItems = append(allGitSubItems, t.toOpenBranchMenuItem(b))
			}
		}
		return allGitSubItems
	}})

	items = append(items, cui.MenuItem{Text: "Live and Deleted Branches", ItemsFunc: func() []cui.MenuItem {
		var allSubItems []cui.MenuItem
		for _, b := range t.vm.GetAllBranches(true) {
			allSubItems = append(allSubItems, t.toOpenBranchMenuItem(b))
		}
		return allSubItems
	}})

	if len(ambiguousBranches) > 0 {
		items = append(items, cui.MenuItem{Text: "Ambiguous Branches", ItemsFunc: func() []cui.MenuItem {
			var allSubItems []cui.MenuItem
			for _, b := range t.vm.GetNotShownAmbiguousBranches() {
				allSubItems = append(allSubItems, t.toOpenBranchMenuItem(b))
			}
			return allSubItems
		}})
	}

	return items
}

func (t *menus) getShowCommitBranchesMenuItems(selectedIndex int) []cui.MenuItem {
	var items []cui.MenuItem
	branches := t.vm.GetCommitBranches(selectedIndex)
	for _, b := range branches {
		items = append(items, t.toOpenBranchMenuItem(b))
	}
	return items
}

func (t *menus) getHideBranchMenuItems() []cui.MenuItem {
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

func (t *menus) getBranchHierarchyMenuItems(commit api.Commit) []cui.MenuItem {
	b := t.vm.repo.Branches[commit.BranchIndex]
	items := []cui.MenuItem{}

	if b.IsSetAsParent {
		txt := fmt.Sprintf("Unset %s as Parent", b.DisplayName)
		items = append(items, cui.MenuItem{Text: txt, Action: func() {
			t.vm.UnsetAsParentBranch(b.Name)
		}})
	} else if commit.IsAmbiguous && len(b.AmbiguousBranchNames) > 0 {
		subItems := lo.Map(b.AmbiguousBranchNames, func(v string, _ int) cui.MenuItem {
			vv := v
			return cui.MenuItem{Text: vv, Action: func() { t.vm.SetAsParentBranch(b.Name, vv) }}
		})

		items = append(items, cui.MenuItem{Text: "Set Ambiguous Branch Parent",
			Items: subItems})
	}

	return items
}

func (t *menus) getSwitchBranchMenuItems(onlyShown bool) []cui.MenuItem {
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

	if onlyShown {
		return items
	}

	items = append(items, cui.MenuSeparator(""))

	items = append(items, cui.MenuItem{Text: "Latest Branches", ItemsFunc: func() []cui.MenuItem {
		var activeSubItems []cui.MenuItem
		for _, b := range t.vm.GetRecentBranches(true) {
			bb := b // closure save
			switchItem := cui.MenuItem{Text: t.branchItemText(b), Action: func() {
				t.vm.SwitchToBranch(bb.Name, bb.DisplayName)
			}}
			activeSubItems = append(activeSubItems, switchItem)
		}
		return activeSubItems
	}})

	items = append(items, cui.MenuItem{Text: "Live Branches", ItemsFunc: func() []cui.MenuItem {
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

	items = append(items, cui.MenuItem{Text: "Live and Deleted Branches", ItemsFunc: func() []cui.MenuItem {
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

func (t *menus) toOpenBranchMenuItem(branch api.Branch) cui.MenuItem {
	text := t.branchItemText(branch)
	if !branch.IsGitBranch {
		// Not a git branch, mark the branch a bit darker
		text = cui.Dark(text)
	}

	return cui.MenuItem{Text: text, Action: func() {
		t.vm.ShowBranch(branch.Name, "")
	}}
}

// func (t *menuService) toSetAsParentBranchMenuItem(branch api.Branch) cui.MenuItem {
// 	text := t.branchItemText(branch)

// 	return cui.MenuItem{Text: text, Action: func() {
// 		t.vm.SetAsParentBranch(branch.Name)
// 	}}
// }

func (t *menus) branchItemText(branch api.Branch) string {
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

func (t *menus) getPushBranchMenuItems() []cui.MenuItem {
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

func (t *menus) getPullBranchMenuItems() []cui.MenuItem {
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

func (t *menus) getMergeMenuItems() []cui.MenuItem {
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

func (t *menus) getFileDiffsMenuItems() []cui.MenuItem {
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

func (t *menus) getDeleteBranchMenuItems() []cui.MenuItem {
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
