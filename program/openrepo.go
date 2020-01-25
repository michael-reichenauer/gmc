package program

import (
	"github.com/michael-reichenauer/gmc/utils/log"
	"github.com/michael-reichenauer/gmc/utils/ui"
	"io/ioutil"
	"os"
	"path/filepath"
)

func (h *MainWindow) GetOpenRepoMenuItem() ui.MenuItem {
	return ui.MenuItem{Text: "Open repo", Title: "Open Repo", Key: "Ctrl-O", SubItems: h.getOpenRepoMenuItems()}
}

func (h *MainWindow) GetOpenRepoMenu() *ui.Menu {
	menu := ui.NewMenu(h.uiHandler, "Open repo")
	menu.AddItems(h.getOpenRepoMenuItems())
	return menu
}

func (h *MainWindow) getOpenRepoMenuItems() []ui.MenuItem {
	return []ui.MenuItem{
		h.getRecentMenuItem(),
		h.getOpenMenuItem(),
	}
}

func (h *MainWindow) getRecentMenuItem() ui.MenuItem {
	var items []ui.MenuItem
	for _, f := range h.configService.GetState().RecentFolders {
		items = append(items, ui.MenuItem{Text: f})
	}
	return ui.MenuItem{Text: "Recent Repos", Title: "Recent Repos", SubItems: items}
}

func (h *MainWindow) getOpenMenuItem() ui.MenuItem {
	home, err := os.UserHomeDir()
	if err != nil {
		panic(log.Fatal(err))
	}

	return ui.MenuItem{Text: "Open Repo", Title: "Open Repo", SubItemsFunc: func() []ui.MenuItem {
		return h.getOpenFolderItems(home)
	}}
}

func (h *MainWindow) getOpenFolderItems(folder string) []ui.MenuItem {
	files, err := ioutil.ReadDir(folder)
	if err != nil {
		log.Fatal(err)
	}

	var items []ui.MenuItem
	for _, f := range files {
		if !f.IsDir() {
			continue
		}
		path := filepath.Join(folder, f.Name())
		items = append(items, ui.MenuItem{Text: f.Name(), SubItemsFunc: func() []ui.MenuItem {
			return h.getOpenFolderItems(path)
		}})
	}
	return items
}
