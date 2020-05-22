package console

import (
	"fmt"
	"github.com/michael-reichenauer/gmc/api"
	"github.com/michael-reichenauer/gmc/utils/cui"
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
	menu := t.ui.NewMenu("")

	menu.Add(cui.MenuItem{Text: "Show Branch", SubItems: t.getOpenBranchMenuItems(currentLineIndex)})
	menu.Add(cui.MenuItem{Text: "Hide Branch", SubItems: t.getCloseBranchMenuItems()})

	menu.Add(cui.SeparatorMenuItem)

	c := t.vm.repo.Commits[currentLineIndex]
	menu.Add(cui.MenuItem{Text: "Commit Diff ...", Key: "Ctrl-D", Action: func() {
		t.vm.showCommitDiff(c.ID)
	}})
	menu.Add(cui.MenuItem{Text: "Commit ...", Key: "Ctrl-S", Action: t.vm.showCommitDialog})
	menu.Add(cui.MenuItem{Text: "Create Branch ...", Key: "Ctrl-B", Action: t.vm.showCreateBranchDialog})
	menu.Add(cui.MenuItem{Text: "Delete Branch", SubItems: t.getDeleteBranchMenuItems()})
	menu.Add(cui.MenuItem{Text: "Push", SubItems: t.getPushBranchMenuItems()})
	menu.Add(cui.MenuItem{Text: "Pull/Update", SubItems: t.getPullBranchMenuItems()})

	menu.Add(cui.MenuItem{Text: "Switch/Checkout", SubItems: t.getSwitchBranchMenuItems()})
	mergeItems, mergeTitle := t.getMergeMenuItems()
	menu.Add(cui.MenuItem{Text: "Merge", Title: mergeTitle, SubItems: mergeItems})

	// menu.Add(t.vm.mainService.RecentReposMenuItem())
	// menu.Add(t.vm.mainService.MainMenuItem())
	return menu
}

func (t *menuService) getShowMoreMenu(selectedIndex int) cui.Menu {
	menu := t.ui.NewMenu("")
	menu.AddItems(t.getOpenBranchMenuItems(selectedIndex))
	menu.Add(cui.MenuItem{Text: "Hide Branch", SubItems: t.getCloseBranchMenuItems()})
	return menu
}

func (t *menuService) getOpenBranchMenuItems(selectedIndex int) []cui.MenuItem {
	var items []cui.MenuItem
	inBranches := t.vm.GetCommitOpenInBranches(selectedIndex)
	for _, b := range inBranches {
		items = append(items, t.toOpenBranchMenuItem(b, "╮"))
	}
	outBranches := t.vm.GetCommitOpenOutBranches(selectedIndex)
	for _, b := range outBranches {
		items = append(items, t.toOpenBranchMenuItem(b, "╭"))
	}

	current, ok := t.vm.CurrentNotShownBranch()
	if ok {
		if nil == funk.Find(append(inBranches, outBranches...), func(b api.Branch) bool {
			return current.DisplayName == b.DisplayName
		}) {
			items = append(items, t.toOpenBranchMenuItem(current, ""))
		}
	}

	if len(items) > 0 {
		items = append(items, cui.SeparatorMenuItem)
	}

	var activeSubItems []cui.MenuItem
	for _, b := range t.vm.GetLatestBranches(true) {
		activeSubItems = append(activeSubItems, t.toOpenBranchMenuItem(b, ""))
	}
	items = append(items, cui.MenuItem{Text: "Latest Branches", SubItems: activeSubItems})

	var allGitSubItems []cui.MenuItem
	for _, b := range t.vm.GetAllBranches(true) {
		if b.IsGitBranch {
			allGitSubItems = append(allGitSubItems, t.toOpenBranchMenuItem(b, ""))
		}
	}
	items = append(items, cui.MenuItem{Text: "All Git Branches", SubItems: allGitSubItems})

	var allSubItems []cui.MenuItem
	for _, b := range t.vm.GetAllBranches(true) {
		allSubItems = append(allSubItems, t.toOpenBranchMenuItem(b, ""))
	}
	items = append(items, cui.MenuItem{Text: "All Branches", SubItems: allSubItems})

	return items
}

func (t *menuService) getCloseBranchMenuItems() []cui.MenuItem {
	var items []cui.MenuItem
	commitBranches := t.vm.GetShownBranches(true)
	for _, b := range commitBranches {
		name := b.Name // closure save
		closeItem := cui.MenuItem{Text: t.branchItemText(b, ""), Action: func() {
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
		bb := b // closure save
		switchItem := cui.MenuItem{Text: t.branchItemText(b, ""), Action: func() {
			t.vm.SwitchToBranch(bb.Name, bb.DisplayName)
		}}
		items = append(items, switchItem)
	}

	var activeSubItems []cui.MenuItem
	for _, b := range t.vm.GetLatestBranches(true) {
		bb := b // closure save
		switchItem := cui.MenuItem{Text: t.branchItemText(b, ""), Action: func() {
			t.vm.SwitchToBranch(bb.Name, bb.DisplayName)
		}}
		activeSubItems = append(activeSubItems, switchItem)
	}
	items = append(items, cui.MenuItem{Text: "Latest Branches", SubItems: activeSubItems})

	var allGitSubItems []cui.MenuItem
	for _, b := range t.vm.GetAllBranches(true) {
		bb := b // closure save
		if b.IsGitBranch {
			switchItem := cui.MenuItem{Text: t.branchItemText(b, ""), Action: func() {
				t.vm.SwitchToBranch(bb.Name, bb.DisplayName)
			}}
			allGitSubItems = append(allGitSubItems, switchItem)
		}
	}
	items = append(items, cui.MenuItem{Text: "All Git Branches", SubItems: allGitSubItems})

	var allSubItems []cui.MenuItem
	for _, b := range t.vm.GetAllBranches(true) {
		bb := b // closure save
		switchItem := cui.MenuItem{Text: t.branchItemText(b, ""), Action: func() {
			t.vm.SwitchToBranch(bb.Name, bb.DisplayName)
		}}
		allSubItems = append(allSubItems, switchItem)
	}
	items = append(items, cui.MenuItem{Text: "All Branches", SubItems: allSubItems})

	return items
}

func (t *menuService) toOpenBranchMenuItem(branch api.Branch, prefix string) cui.MenuItem {
	return cui.MenuItem{Text: t.branchItemText(branch, prefix), Action: func() {
		t.vm.ShowBranch(branch.Name)
	}}
}

func (t *menuService) branchItemText(branch api.Branch, prefix string) string {
	if prefix == "" {
		prefix = " "
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
		pushItem := cui.MenuItem{Text: t.branchItemText(current, ""), Key: "Ctrl-P", Action: func() {
			t.vm.PushBranch(current.Name)
		}}
		items = append(items, pushItem)
	}
	return items
}

func (t *menuService) getPullBranchMenuItems() []cui.MenuItem {
	var items []cui.MenuItem
	current, ok := t.vm.CurrentBranch()
	if ok && current.HasRemoteOnly {
		pushItem := cui.MenuItem{Text: t.branchItemText(current, ""), Key: "Ctrl-U", Action: func() {
			t.vm.PullCurrentBranch()
		}}
		items = append(items, pushItem)
	}
	return items
}

func (t *menuService) getMergeMenuItems() ([]cui.MenuItem, string) {
	current, ok := t.vm.CurrentBranch()
	if !ok {
		return nil, ""
	}
	var items []cui.MenuItem
	commitBranches := t.vm.GetShownBranches(false)
	for _, b := range commitBranches {
		name := b.Name // closure save
		if b.DisplayName == current.DisplayName {
			continue
		}
		item := cui.MenuItem{Text: t.branchItemText(b, ""), Action: func() {
			t.vm.MergeFromBranch(name)
		}}
		items = append(items, item)
	}
	return items, fmt.Sprintf("Merge to: %s", current.DisplayName)
}

func (t *menuService) getDeleteBranchMenuItems() []cui.MenuItem {
	var items []cui.MenuItem
	branches := t.vm.GetAllBranches(false)
	for _, b := range branches {
		if !b.IsGitBranch || b.DisplayName == "master" || b.IsCurrent {
			continue
		}
		name := b.Name // closure save
		item := cui.MenuItem{Text: t.branchItemText(b, ""), Action: func() {
			t.vm.DeleteBranch(name)
		}}
		items = append(items, item)
	}
	return items

}
