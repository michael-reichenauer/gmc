package console

import (
	"fmt"

	"github.com/michael-reichenauer/gmc/api"
	"github.com/michael-reichenauer/gmc/utils/cui"
	"github.com/michael-reichenauer/gmc/utils/git"
	"github.com/michael-reichenauer/gmc/utils/linq"
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
	ambiguousBranches := t.vm.GetAmbiguousBranches()
	items := []cui.MenuItem{}

	current, ok := t.vm.CurrentBranch()
	if ok {
		items = append(items, t.toShowBranchMenuItem(current))
	}
	items = append(items, linq.Map(t.vm.GetCommitBranches(selectedIndex), t.toShowBranchMenuItem)...)

	items = append(items, cui.MenuSeparator(""))

	items = append(items, cui.MenuItem{Text: "Recent Branches", ItemsFunc: func() []cui.MenuItem {
		return linq.Map(t.vm.GetRecentBranches(), t.toShowBranchMenuItem)
	}})
	items = append(items, cui.MenuItem{Text: "Live Branches", ItemsFunc: func() []cui.MenuItem {
		return linq.Map(t.vm.GetAllGitBranches(), t.toShowBranchMenuItem)
	}})
	items = append(items, cui.MenuItem{Text: "Live and Deleted Branches", ItemsFunc: func() []cui.MenuItem {
		return linq.Map(t.vm.GetAllBranches(), t.toShowBranchMenuItem)
	}})
	if len(ambiguousBranches) > 0 {
		items = append(items, cui.MenuItem{Text: "Ambiguous Branches", ItemsFunc: func() []cui.MenuItem {
			return linq.Map(t.vm.GetAmbiguousBranches(), t.toShowBranchMenuItem)
		}})
	}

	return items
}

func (t *menus) getHideBranchMenuItems() []cui.MenuItem {
	return linq.Map(t.vm.GetShownBranches(true), t.toHideBranchMenuItem)
}

func (t *menus) toHideBranchMenuItem(branch api.Branch) cui.MenuItem {
	return cui.MenuItem{Text: t.branchItemText(branch), Action: func() {
		t.vm.HideBranch(branch.Name)
	}}
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
		items = append(items, cui.MenuItem{Text: "Set Ambiguous Branch Parent", Items: subItems})
	}

	return items
}

func (t *menus) toSwitchBranchMenuItem(branch api.Branch) cui.MenuItem {
	return cui.MenuItem{Text: t.branchItemText(branch), Action: func() {
		t.vm.SwitchToBranch(branch.Name, branch.DisplayName)
	}}
}

func (t *menus) isNotCurrentBranch(branch api.Branch) bool {
	return !branch.IsCurrent
}

func (t *menus) getSwitchBranchMenuItems(onlyShown bool) []cui.MenuItem {
	var items []cui.MenuItem

	items = append(items, linq.FilterMap(t.vm.GetShownBranches(false),
		t.isNotCurrentBranch, t.toSwitchBranchMenuItem)...)

	if onlyShown {
		return items
	}

	items = append(items, cui.MenuSeparator(""))

	items = append(items, cui.MenuItem{Text: "Recent Branches", ItemsFunc: func() []cui.MenuItem {
		return linq.FilterMap(t.vm.GetRecentBranches(),
			t.isNotCurrentBranch, t.toSwitchBranchMenuItem)
	}})

	items = append(items, cui.MenuItem{Text: "Live Branches", ItemsFunc: func() []cui.MenuItem {
		return linq.Map(t.vm.GetAllGitBranches(), t.toSwitchBranchMenuItem)
	}})

	items = append(items, cui.MenuItem{Text: "Live and Deleted Branches", ItemsFunc: func() []cui.MenuItem {
		return linq.Map(t.vm.GetAllBranches(), t.toSwitchBranchMenuItem)
	}})

	return items
}

func (t *menus) toShowBranchMenuItem(branch api.Branch) cui.MenuItem {
	text := t.branchItemText(branch)
	if !branch.IsGitBranch {
		// Not a git branch, mark the branch a bit darker
		text = cui.Dark(text)
	}

	return cui.MenuItem{Text: text, Action: func() {
		t.vm.ShowBranch(branch.Name, "")
	}}
}

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
	branches := t.vm.GetAllBranches()
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
