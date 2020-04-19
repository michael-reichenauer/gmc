package repoview

import (
	"github.com/michael-reichenauer/gmc/repoview/viewmodel"
	"github.com/michael-reichenauer/gmc/utils/ui"
	"github.com/thoas/go-funk"
)

type menuService struct {
	ui *ui.UI
	vm *repoVM
}

func newMenuService(ui *ui.UI, vm *repoVM) *menuService {
	return &menuService{ui: ui, vm: vm}
}

func (t *menuService) getContextMenu(currentLineIndex int) *ui.Menu {
	menu := t.ui.NewMenu("")

	showItems := t.getOpenBranchMenuItems()
	menu.Add(ui.MenuItem{Text: "Show Branch", SubItems: showItems})

	hideItems := t.GetCloseBranchMenuItems()
	menu.Add(ui.MenuItem{Text: "Hide Branch", SubItems: hideItems})

	menu.Add(ui.SeparatorMenuItem)
	c := t.vm.repo.Commits[currentLineIndex]
	menu.Add(ui.MenuItem{Text: "Commit Diff ...", Key: "Ctrl-D", Action: func() {
		t.vm.showCommitDiff(c.ID)
	}})
	menu.Add(ui.MenuItem{Text: "Commit ...", Key: "Ctrl-Space", Action: func() {
		t.vm.commit()
	}})

	pushItems := t.getPushBranchMenuItems()
	if pushItems != nil {
		menu.Add(ui.MenuItem{Text: "Push", SubItems: pushItems})
	}

	switchItems := t.GetSwitchBranchMenuItems()
	menu.Add(ui.MenuItem{Text: "Switch/Checkout", SubItems: switchItems})

	menu.Add(t.vm.mainService.RecentReposMenuItem())
	menu.Add(t.vm.mainService.MainMenuItem())
	return menu
}

func (t *menuService) getOpenBranchMenuItems() []ui.MenuItem {
	branches := t.vm.GetCommitOpenBranches()

	current, ok := t.vm.CurrentNotShownBranch()
	if ok {
		if nil == funk.Find(branches, func(b viewmodel.Branch) bool {
			return current.DisplayName == b.DisplayName
		}) {
			branches = append(branches, current)
		}
	}

	var items []ui.MenuItem
	for _, b := range branches {
		items = append(items, t.toOpenBranchMenuItem(b))
	}

	if len(items) > 0 {
		items = append(items, ui.SeparatorMenuItem)
	}

	var activeSubItems []ui.MenuItem
	for _, b := range t.vm.GetActiveBranches() {
		activeSubItems = append(activeSubItems, t.toOpenBranchMenuItem(b))
	}
	items = append(items, ui.MenuItem{Text: "Active Branches", SubItems: activeSubItems})

	var allGitSubItems []ui.MenuItem
	for _, b := range t.vm.GetAllBranches() {
		if b.IsGitBranch {
			allGitSubItems = append(allGitSubItems, t.toOpenBranchMenuItem(b))
		}
	}
	items = append(items, ui.MenuItem{Text: "All Git Branches", SubItems: allGitSubItems})

	var allSubItems []ui.MenuItem
	for _, b := range t.vm.GetAllBranches() {
		allSubItems = append(allSubItems, t.toOpenBranchMenuItem(b))
	}
	items = append(items, ui.MenuItem{Text: "All Branches", SubItems: allSubItems})

	return items

}

func (t *menuService) GetCloseBranchMenuItems() []ui.MenuItem {
	var items []ui.MenuItem
	commitBranches := t.vm.GetShownBranches(true)
	for _, b := range commitBranches {
		closeItem := ui.MenuItem{Text: t.branchItemText(b), Action: func() {
			t.vm.HideBranch(b.Name)
		}}
		items = append(items, closeItem)
	}
	return items
}

func (t *menuService) GetSwitchBranchMenuItems() []ui.MenuItem {
	var items []ui.MenuItem
	commitBranches := t.vm.GetShownBranches(false)
	for _, b := range commitBranches {
		switchItem := ui.MenuItem{Text: t.branchItemText(b), Action: func() {
			t.vm.SwitchToBranch(b.Name)
		}}
		items = append(items, switchItem)
	}
	return items
}

func (t *menuService) toOpenBranchMenuItem(branch viewmodel.Branch) ui.MenuItem {
	return ui.MenuItem{Text: t.branchItemText(branch), Action: func() {
		t.vm.ShowBranch(branch.Name)
	}}
}

func (t *menuService) branchItemText(branch viewmodel.Branch) string {
	if branch.IsCurrent {
		return "●" + branch.DisplayName
	} else {
		return " " + branch.DisplayName
	}
}

func (t *menuService) getPushBranchMenuItems() []ui.MenuItem {
	var items []ui.MenuItem
	current, ok := t.vm.CurrentBranch()
	if ok && current.HasLocalOnly {
		pushItem := ui.MenuItem{Text: t.branchItemText(current), Action: func() {
			t.vm.PushBranch(current.Name)
		}}
		items = append(items, pushItem)
	}
	return items
}
