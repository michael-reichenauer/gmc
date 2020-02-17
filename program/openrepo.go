package program

import (
	"github.com/michael-reichenauer/gmc/common/config"
	"github.com/michael-reichenauer/gmc/utils"
	"github.com/michael-reichenauer/gmc/utils/git"
	"github.com/michael-reichenauer/gmc/utils/log"
	"github.com/michael-reichenauer/gmc/utils/ui"
	"io/ioutil"
	"path/filepath"
	"sort"
	"strings"
)

func (h *MainWindow) GetOpenRepoMenuItem() ui.MenuItem {
	return ui.MenuItem{Text: "Open repo", Title: "Open Repo", Key: "Ctrl-O",
		SubItems: h.OpenRepoMenuItems()}
}

func (h *MainWindow) GetOpenRepoMenu() *ui.Menu {
	menu := ui.NewMenu(h.uiHandler, "Open repo")
	menu.AddItems(h.OpenRepoMenuItems())
	return menu
}

func (h *MainWindow) GetStartMenu() *ui.Menu {
	menu := ui.NewMenu(h.uiHandler, "Open repo")
	menu.AddItems(h.OpenRepoMenuItems2())
	return menu
}

func (h *MainWindow) OpenRepo(folderPath string) {
	log.Infof("Opening %q ...", folderPath)
	workingFolder, err := git.WorkingFolderRoot(folderPath)
	if err != nil {
		log.Warnf("No working folder %q", folderPath)
		openMenu := h.GetStartMenu()
		openMenu.Show(3, 1)
		return
	}
	log.Infof("Got repo %q ...", workingFolder)

	err = h.repoViewModelService.OpenRepo(workingFolder)
	if err != nil {
		log.Warnf("Failed to open repo %q", workingFolder)
		openMenu := h.GetOpenRepoMenu()
		openMenu.Show(1, 1)
		return
	}
	parent := filepath.Dir(workingFolder)

	h.configService.SetState(func(s *config.State) {
		s.RecentFolders = utils.RecentItems(s.RecentFolders, workingFolder, 10)
		s.RecentParentFolders = utils.RecentItems(s.RecentParentFolders, parent, 5)
	})
}

func (h *MainWindow) OpenRepoMenuItems() []ui.MenuItem {
	return []ui.MenuItem{
		h.RecentReposMenuItem(),
		h.getOpenMenuItem(),
	}
}

func (h *MainWindow) OpenRepoMenuItems2() []ui.MenuItem {
	items := h.getRecentMenuItems()
	if len(items) > 0 {
		items = append(items, ui.SeparatorMenuItem)
	}
	items = append(items, h.getOpenMenuItem())
	return items
}

func (h *MainWindow) RecentReposMenuItem() ui.MenuItem {
	return ui.MenuItem{Text: "Recent Repos", Title: "Recent Repos", SubItems: h.getRecentMenuItems()}
}

func (h *MainWindow) getRecentMenuItems() []ui.MenuItem {
	var items []ui.MenuItem
	for _, f := range h.configService.GetState().RecentFolders {
		path := f
		items = append(items, ui.MenuItem{Text: path, Action: func() { h.OpenRepo(path) }})
	}
	return items
}

func (h *MainWindow) getOpenMenuItem() ui.MenuItem {
	paths := h.configService.GetState().RecentParentFolders
	paths = append(paths, utils.GetVolumes()...)
	if len(paths) == 1 {
		return ui.MenuItem{Text: "Open", Title: paths[0], SubItemsFunc: func() []ui.MenuItem {
			return h.getFolderItems(paths[0], func(folder string) { h.OpenRepo(folder) })
		}}
	}

	var items []ui.MenuItem
	for _, p := range paths {
		path := p
		items = append(items,
			ui.MenuItem{Text: path, Title: path, SubItemsFunc: func() []ui.MenuItem {
				return h.getFolderItems(path, func(folder string) { h.OpenRepo(folder) })
			}})
	}
	return ui.MenuItem{Text: "Open Repo", SubItems: items}
}

func (h *MainWindow) getFolderItems(folder string, action func(f string)) []ui.MenuItem {
	files, err := ioutil.ReadDir(folder)
	if err != nil {
		// Folder not readable, might be e.g. access denied
		return nil
	}

	var items []ui.MenuItem

	for _, f := range files {
		if !f.IsDir() || f.Name() == "$RECYCLE.BIN" {
			continue
		}
		path := filepath.Join(folder, f.Name())
		items = append(items, ui.MenuItem{
			Text:   f.Name(),
			Title:  path,
			Action: func() { action(path) },
			SubItemsFunc: func() []ui.MenuItem {
				return h.getFolderItems(path, action)
			},
			ReuseBounds: true,
		})
	}
	sort.SliceStable(items, func(l, r int) bool {
		return -1 == strings.Compare(strings.ToLower(items[l].Text), strings.ToLower(items[r].Text))
	})
	return items
}
