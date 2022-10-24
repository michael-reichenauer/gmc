package console

import (
	"fmt"
	"path"
	"strings"

	"github.com/michael-reichenauer/gmc/api"
	"github.com/michael-reichenauer/gmc/utils/async"
	"github.com/michael-reichenauer/gmc/utils/cui"
	"github.com/michael-reichenauer/gmc/utils/log"
)

type MainWindow struct {
	ui  cui.UI
	api api.Api
}

func NewMainWindow(ui cui.UI) *MainWindow {
	return &MainWindow{ui: ui}
}

func (t *MainWindow) Show(api api.Api, path string) {
	t.api = api
	t.showRepo(path)
}

func (t *MainWindow) showRepo(path string) {
	progress := t.ui.ShowProgress("Opening repo:\n%s", path)

	async.RunRE(func() (string, error) {
		var repoId string
		err := t.api.OpenRepo(path, &repoId)
		return repoId, err
	}).Then(func(repoId string) {
		progress.Close()
		repoView := NewRepoView(t.ui, t.api, t, repoId)
		repoView.Show()
	}).Catch(func(err error) {
		if path != "" {
			log.Warnf("Failed to open %q, %v", path, err)
			msgBox := t.ui.MessageBox("Error !", cui.Red(fmt.Sprintf("Failed to show repo for:\n%s\nError: %v", path, err)))
			msgBox.OnClose = func() { t.ui.Post(func() { t.showOpenRepoMenu() }) }
			msgBox.Show()
		} else {
			t.showOpenRepoMenu()
		}
	})
}

func (t *MainWindow) Close() {
}

func (t *MainWindow) showOpenRepoMenu() {
	menu := t.ui.NewMenu("Open repo")
	menu.OnClose(func() { t.ui.Quit() })

	items := t.OpenRepoMenuItems()
	menu.AddItems(items)
	menu.Show(3, 1)
}

func (t *MainWindow) OpenRepoMenuItems() []cui.MenuItem {
	var items []cui.MenuItem

	var recentDirs []string
	err := t.api.GetRecentWorkingDirs(api.NilArg, &recentDirs)
	if err != nil {
		return items
	}
	if len(recentDirs) > 0 {
		items = append(items, t.getRecentRepoMenuItems(recentDirs)...)
		items = append(items, cui.MenuSeparator(""))
	}

	var paths []string
	err = t.api.GetSubDirs("", &paths)
	if err != nil {
		return items
	}

	openItemsFunc := func() []cui.MenuItem {
		return t.getDirItems(paths, func(path string) { t.showRepo(path) })
	}

	items = append(items, cui.MenuItem{Text: "Browse Folders", Title: "Browse", ItemsFunc: openItemsFunc})
	return items
}

func (t *MainWindow) getRecentRepoMenuItems(recentDirs []string) []cui.MenuItem {
	var items []cui.MenuItem
	for _, f := range recentDirs {
		path := f
		items = append(items, cui.MenuItem{Text: path, Action: func() { t.showRepo(path) }})
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
			ItemsFunc: func() []cui.MenuItem {
				var dirs []string
				err := t.api.GetSubDirs(path, &dirs)
				if err != nil {
					log.Warnf("Failed to list %q folder, %v", path, err)
				}
				return t.getDirItems(dirs, action)
			},
			ReuseBounds: true,
		})
	}
	return items
}
