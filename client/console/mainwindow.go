package console

import (
	"fmt"
	"path"
	"strings"

	"github.com/michael-reichenauer/gmc/api"
	"github.com/michael-reichenauer/gmc/common/config"
	"github.com/michael-reichenauer/gmc/utils/cui"
	"github.com/michael-reichenauer/gmc/utils/log"
)

type MainWindow struct {
	ui            cui.UI
	api           api.Api
	configService *config.Service
}

func NewMainWindow(ui cui.UI, configService *config.Service) *MainWindow {
	return &MainWindow{ui: ui, configService: configService}
}

func (t *MainWindow) Show(api api.Api, path string) {
	t.api = api
	t.ShowRepo(path)
}

func (t *MainWindow) ShowRepo(path string) {
	progress := t.ui.ShowProgress("Opening repo:\n%s", path)

	t.api.OpenRepo(path).
		Then(func(repoId string) {
			progress.Close()
			repoView := NewRepoView(t.ui, t.api, t, t.configService, repoId)
			repoView.Show()
		}).
		Catch(func(err error) {
			progress.Close()
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

	recentDirs, err := t.api.GetRecentWorkingDirs()
	if err != nil {
		return items
	}
	if len(recentDirs) > 0 {
		items = append(items, t.getRecentRepoMenuItems(recentDirs)...)
		items = append(items, cui.MenuSeparator(""))
	}

	paths, err := t.api.GetSubDirs("")
	if err != nil {
		return items
	}

	openItemsFunc := func() []cui.MenuItem {
		return t.getDirItems(paths, func(path string) { t.ShowRepo(path) })
	}

	items = append(items, cui.MenuItem{Text: "Browse Folders", Title: "Browse", ItemsFunc: openItemsFunc})
	return items
}

func (t *MainWindow) getRecentRepoMenuItems(recentDirs []string) []cui.MenuItem {
	var items []cui.MenuItem
	for _, f := range recentDirs {
		path := f
		items = append(items, cui.MenuItem{Text: path, Action: func() { t.ShowRepo(path) }})
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
				dirs, err := t.api.GetSubDirs(path)
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
