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
	GetMainMenu(currentLineIndex int) cui.Menu
	GetShowBranchesMenu(selectedIndex int) cui.Menu
	GetHideBranchesMenu() cui.Menu
}

type menus struct {
	ui cui.UI
	vm *repoVM
}

func newMenus(ui cui.UI, vm *repoVM) Menus {
	return &menus{ui: ui, vm: vm}
}

func (t *menus) GetMainMenu(currentLineIndex int) cui.Menu {
	menu := t.ui.NewMenu("Main Menu")
	menu.AddItems(t.getMainMenuItems(currentLineIndex))
	return menu
}

func (t *menus) GetShowBranchesMenu(selectedIndex int) cui.Menu {
	menu := t.ui.NewMenu("Branches")
	menu.Add(cui.MenuSeparator("Show"))

	menu.AddItems(t.getShowImmediateBranchesMenuItems(selectedIndex))

	menu.Add(cui.MenuSeparator("Switch to"))
	menu.AddItems(t.getSwitchBranchMenuItems(true))

	menu.Add(cui.MenuSeparator("More"))
	menu.Add(cui.MenuItem{Text: "Show Branch", Title: "Show More Branches", Key: "->", ItemsFunc: func() []cui.MenuItem {
		return t.getShowBranchesSubSubMenuItems(selectedIndex)
	}})

	menu.Add(cui.MenuItem{Text: "Main Menu", Title: "Main Menu", Key: "M", Items: t.getMainMenuItems(selectedIndex)})
	return menu
}

func (t *menus) GetHideBranchesMenu() cui.Menu {
	menu := t.ui.NewMenu("Hide Branch")
	menu.AddItems(t.getHideBranchMenuItems())
	return menu
}

func (t *menus) getMainMenuItems(currentLineIndex int) []cui.MenuItem {
	c := t.vm.repo.Commits[currentLineIndex]
	items := []cui.MenuItem{}

	// Commit items
	items = append(items, cui.MenuSeparator(fmt.Sprintf("Commit: %s", c.SID)))
	items = append(items, cui.MenuItem{Text: "Toggle Details ...", Key: "Enter", Action: t.vm.repoViewer.ShowCommitDetails})
	if c.ID == git.UncommittedID {
		items = append(items, cui.MenuItem{Text: "Commit ...", Key: "C", Action: t.vm.showCommitDialog})
	}
	items = append(items, cui.MenuItem{Text: "Commit Diff ...", Key: "D", Action: func() { t.vm.showCommitDiff(c.ID) }})
	items = append(items, cui.MenuItem{Text: "Undo/Restore", Title: "Undo", ItemsFunc: t.getUndoMenuItems})

	// Branches items
	items = append(items, cui.MenuSeparator("Branches"))
	items = append(items, cui.MenuItem{Text: "Show Branch", Title: "Show Branch", Key: "->", ItemsFunc: func() []cui.MenuItem {
		return t.getShowBranchesSubMenuItems(currentLineIndex)
	}})
	items = append(items, cui.MenuItem{Text: "Hide Branch", Title: "Hide Branch", Key: "<-", ItemsFunc: t.getHideBranchMenuItems})
	items = append(items, cui.MenuItem{Text: "Switch/Checkout", Title: "Switch To", ItemsFunc: func() []cui.MenuItem {
		return t.getSwitchBranchMenuItems(false)
	}})
	items = append(items, cui.MenuItem{Text: "Push", Title: "Push", ItemsFunc: t.getPushBranchMenuItems})
	items = append(items, cui.MenuItem{Text: "Update/Pull", Title: "Update", ItemsFunc: t.getPullBranchMenuItems})
	items = append(items, cui.MenuItem{Text: "Merge", Title: fmt.Sprintf("Merge Into: %s", t.vm.repo.CurrentBranchName),
		ItemsFunc: t.getMergeMenuItems})
	items = append(items, cui.MenuItem{Text: "MergeSquash", Title: fmt.Sprintf("MergeSquash Into: %s", t.vm.repo.CurrentBranchName),
		ItemsFunc: t.getMergeSquashMenuItems})
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
	items = append(items, cui.MenuItem{Text: "Clone Repo ...", Title: "Clone", Action: t.vm.showCloneDialog})
	items = append(items, cui.MenuItem{Text: "Help ...", Key: "H", Action: func() { ShowHelpDlg(t.ui) }})

	items = append(items, cui.MenuItem{Text: "About ...", Action: func() { ShowAboutDlg(t.ui) }})
	items = append(items, cui.MenuItem{Text: "Quit", Key: "Esc", Action: func() { t.ui.Quit() }})

	return items
}

func (t *menus) getShowBranchesSubMenuItems(selectedIndex int) []cui.MenuItem {
	items := t.getShowImmediateBranchesMenuItems(selectedIndex)

	if len(items) > 0 {
		items = append(items, cui.MenuSeparator(""))
	}

	items = append(items, t.getShowBranchesSubSubMenuItems(selectedIndex)...)

	return items
}

func (t *menus) getShowImmediateBranchesMenuItems(selectedIndex int) []cui.MenuItem {
	items := []cui.MenuItem{}

	items = append(items, linq.Map(t.vm.GetCommitBranches(selectedIndex), t.toShowBranchMenuItem)...)

	current, ok := t.vm.CurrentBranch()
	if ok && !linq.ContainsBy(t.vm.GetShownBranches(false), func(v api.Branch) bool { return v.IsCurrent }) {
		items = append(items, t.toShowBranchMenuItem(current))
	}
	if len(items) > 0 {
		items = append([]cui.MenuItem{cui.MenuSeparator("Show")}, items...)
	}

	shownItems := linq.Map(t.vm.GetShownBranches(true), t.toShowBranchMenuItem)
	items = append(items, shownItems...)

	return items
}

func (t *menus) getShowBranchesSubSubMenuItems(selectedIndex int) []cui.MenuItem {
	items := []cui.MenuItem{}
	ambiguousBranches := t.vm.GetAmbiguousBranches()

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
	}

	if commit.IsAmbiguous && len(b.AmbiguousBranchNames) > 0 {
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
		prefix = "╯"
	}
	if branch.IsCurrent {
		return prefix + "●" + branch.DisplayName
	} else {
		return prefix + " " + branch.DisplayName
	}
}

func (t *menus) getUndoMenuItems() []cui.MenuItem {
	var items []cui.MenuItem

	// Add current branch if it has commits that can be pushed
	current, ok := t.vm.CurrentBranch()
	if ok {
		if current.TipID == git.UncommittedID {
			items = append(items, cui.MenuItem{Text: "Undo/Restore Uncommitted File",
				Title:     "Undo/Restore File",
				ItemsFunc: t.getUncommittedFilesMenuItems})
			items = append(items, cui.MenuSeparator(""))
			items = append(items, cui.MenuItem{Text: "Undo/Restore all Uncommitted Changes",
				Action: func() { t.vm.UndoAllUncommittedChanges() }})
		} else {
			c, ok := linq.Find(t.vm.repo.Commits, func(v api.Commit) bool { return v.ID == current.TipID })
			if ok {
				if c.IsLocalOnly {
					items = append(items, cui.MenuItem{Text: "Uncommit Last Commit", Action: func() {
						t.vm.UncommitLastCommit()
					}})
				}
			}
		}
	}

	if current.TipID != git.UncommittedID {
		c := t.vm.repo.Commits[t.vm.currentIndex]
		txt := fmt.Sprintf("Undo Commit %s", c.SID)
		items = append(items, cui.MenuItem{Text: txt, Action: func() {
			t.vm.UndoCommit(c.ID)
		}})
	}

	items = append(items, cui.MenuItem{Text: "Clean/Restore Working Folder", Action: func() {
		t.vm.CleanWorkingFolder()
	}})

	return items
}

func (t *menus) getUncommittedFilesMenuItems() []cui.MenuItem {
	files := t.vm.GetUncommittedFiles()

	return linq.Map(files, func(v string) cui.MenuItem {
		return cui.MenuItem{Text: v, Action: func() { t.vm.UndoUncommittedFileChanges(v) }}
	})
}

func (t *menus) getPushBranchMenuItems() []cui.MenuItem {
	var items []cui.MenuItem

	// Add current branch if it has commits that can be pushed
	current, ok := t.vm.CurrentBranch()
	if ok && current.HasLocalOnly {
		pushItem := t.toPushBranchMenuItem(current)
		pushItem.Key = "P"
		items = append(items, pushItem)
	}

	// Add other branches if they have push commits (but only if the can be pushed cleanly)
	otherItems := linq.FilterMap(t.vm.repo.Branches,
		func(v api.Branch) bool { return !v.IsCurrent && !v.IsRemote && v.HasLocalOnly && !v.HasRemoteOnly },
		t.toPushBranchMenuItem)

	// Add separator between current and other branches
	if len(items) > 0 && len(otherItems) > 0 {
		items = append(items, cui.MenuSeparator(""))
	}

	items = append(items, otherItems...)

	return items
}

func (t *menus) getPullBranchMenuItems() []cui.MenuItem {
	var items []cui.MenuItem

	// Add current branch if it has commits that can be pulled
	current, ok := t.vm.CurrentBranch()
	if ok && current.HasRemoteOnly {
		pushItem := t.toPullCurrentBranchMenuItem(current)
		pushItem.Key = "U"
		items = append(items, pushItem)
	}

	// Add other branches if they have pull commits (but only if the can be pulled cleanly)
	otherItems := linq.FilterMap(t.vm.repo.Branches,
		func(b api.Branch) bool { return !b.IsCurrent && b.IsRemote && b.HasRemoteOnly && !b.HasLocalOnly },
		t.toPullBranchMenuItem)

	// Add separator between current and other branches
	if len(items) > 0 && len(otherItems) > 0 {
		items = append(items, cui.MenuSeparator(""))
	}

	items = append(items, otherItems...)
	return items
}

func (t *menus) toPushBranchMenuItem(branch api.Branch) cui.MenuItem {
	return cui.MenuItem{Text: t.branchItemText(branch), Action: func() {
		t.vm.PushBranch(branch.DisplayName)
	}}
}

func (t *menus) toPullBranchMenuItem(branch api.Branch) cui.MenuItem {
	return cui.MenuItem{Text: t.branchItemText(branch), Action: func() {
		t.vm.PullBranch(branch.DisplayName)
	}}
}

func (t *menus) toPullCurrentBranchMenuItem(branch api.Branch) cui.MenuItem {
	return cui.MenuItem{Text: t.branchItemText(branch), Action: func() {
		t.vm.PullCurrentBranch()
	}}
}

func (t *menus) getMergeMenuItems() []cui.MenuItem {
	current, ok := t.vm.CurrentBranch()
	if !ok {
		// No current branch (detached branch)
		return nil
	}

	return linq.FilterMap(t.vm.GetShownBranches(false),
		func(b api.Branch) bool { return b.DisplayName != current.DisplayName },
		func(b api.Branch) cui.MenuItem {
			return cui.MenuItem{Text: t.branchItemText(b), Action: func() {
				t.vm.MergeFromBranch(b.Name)
			}}
		})
}

func (t *menus) getMergeSquashMenuItems() []cui.MenuItem {
	current, ok := t.vm.CurrentBranch()
	if !ok {
		// No current branch (detached branch)
		return nil
	}

	return linq.FilterMap(t.vm.GetShownBranches(false),
		func(b api.Branch) bool { return b.DisplayName != current.DisplayName },
		func(b api.Branch) cui.MenuItem {
			return cui.MenuItem{Text: t.branchItemText(b), Action: func() {
				t.vm.MergeSquashFromBranch(b.Name)
			}}
		})
}

func (t *menus) getFileDiffsMenuItems() []cui.MenuItem {
	c := t.vm.repo.Commits[t.vm.currentIndex]
	ref := c.ID
	if c.ID == git.UncommittedID {
		// For uncommitted changes, the ref is the current branch
		cb, ok := t.vm.CurrentBranch()
		if !ok {
			return []cui.MenuItem{}
		}
		ref = cb.Name
	}

	return linq.Map(t.vm.GetFiles(ref),
		func(v string) cui.MenuItem {
			return cui.MenuItem{Text: v, Action: func() { t.vm.showFileDiff(v) }}
		})
}

func (t *menus) getDeleteBranchMenuItems() []cui.MenuItem {
	return linq.FilterMap(t.vm.GetAllBranches(),
		func(b api.Branch) bool { return b.IsGitBranch && !b.IsMainBranch && !b.IsCurrent },
		func(b api.Branch) cui.MenuItem {
			return cui.MenuItem{Text: t.branchItemText(b), Action: func() {
				t.vm.DeleteBranch(b.Name, false)
			}}
		})
}
