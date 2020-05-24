package console

import (
	"github.com/michael-reichenauer/gmc/api"
	"github.com/michael-reichenauer/gmc/utils/cui"
	"path"
	"strings"
)

type MainWindow struct {
	ui  cui.UI
	api api.Api
}

func NewMainWindow(ui cui.UI, api api.Api) *MainWindow {
	return &MainWindow{ui: ui, api: api}
}

func (t *MainWindow) Show(path string) {
	err := t.api.OpenRepo(path)
	if err != nil {
		if path != "" {
			t.ui.ShowErrorMessageBox("Failed to show repo for:\n%s\nError: %v", path, err)
		}
		t.showOpenRepoMenu()
		return
	}

	repoView := NewRepoView(t.ui, t.api)
	repoView.Show()
}

func (t *MainWindow) showOpenRepoMenu() {
	menu := t.ui.NewMenu("Open repo")

	recentDirs, err := t.api.GetRecentWorkingDirs()
	if err != nil {
		t.ui.ShowErrorMessageBox("Failed to get recent dirs,\nError: %v", err)
	}
	if len(recentDirs) > 0 {
		menu.Add(t.getRecentRepoMenuItems(recentDirs)...)
		menu.Add(cui.SeparatorMenuItem)
	}

	paths, err := t.api.GetSubDirs("")
	if err != nil {
		t.ui.ShowErrorMessageBox("Failed to list of folders to open,\nError: %v", err)
		return
	}

	openItemsFunc := func() []cui.MenuItem {
		return t.getDirItems(paths, func(path string) { t.Show(path) })
	}
	menu.Add(cui.MenuItem{Text: "Open Repo", SubItemsFunc: openItemsFunc})

	menu.Show(3, 1)
}

func (t *MainWindow) getRecentRepoMenuItems(recentDirs []string) []cui.MenuItem {
	var items []cui.MenuItem
	for _, f := range recentDirs {
		path := f
		items = append(items, cui.MenuItem{Text: path, Action: func() { t.Show(path) }})
	}
	return items
}

func (t *MainWindow) getDirItems(paths []string, action func(f string)) []cui.MenuItem {
	var items []cui.MenuItem

	for _, p := range paths {
		name := path.Base(strings.ReplaceAll(p, "\\", "/"))
		path := p
		items = append(items, cui.MenuItem{
			Text:   name,
			Title:  path,
			Action: func() { action(path) },
			SubItemsFunc: func() []cui.MenuItem {
				dirs, err := t.api.GetSubDirs(path)
				if err != nil {
					t.ui.ShowErrorMessageBox("Failed to list of folders to open,\nError: %v", err)
					return nil
				}
				return t.getDirItems(dirs, action)
			},
			ReuseBounds: true,
		})
	}
	return items
}

//
// func (t *MainWindow) MainMenuItem() cui.MenuItem {
// 	var subItems []cui.MenuItem
// 	subItems = append(subItems, t.OpenRepoMenuItems()...)
// 	subItems = append(subItems, cui.MenuItem{Text: "About", Action: t.showAbout})
// 	menuItem := cui.MenuItem{
// 		Text:     "Main Menu",
// 		Title:    "Main Menu",
// 		SubItems: subItems,
// 	}
//
// 	return menuItem
// }

// func (t *MainWindow) showAbout() {
// 	t.ui.ShowMessageBox("About", "gmc version %s\n%s", t.configService.ProgramVersion, git.Version())
// }
