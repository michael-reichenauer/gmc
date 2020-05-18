package program

import (
	"github.com/michael-reichenauer/gmc/common/config"
	"github.com/michael-reichenauer/gmc/cviews"
	"github.com/michael-reichenauer/gmc/utils"
	"github.com/michael-reichenauer/gmc/utils/cui"
	"github.com/michael-reichenauer/gmc/utils/git"
	"github.com/michael-reichenauer/gmc/utils/log"
	"io/ioutil"
	"path/filepath"
	"sort"
	"strings"
)

func (h *MainWindow) GetOpenRepoMenuItem() cui.MenuItem {
	return cui.MenuItem{Text: "Open repo", Title: "Open Repo", Key: "Ctrl-O",
		SubItems: h.OpenRepoMenuItems()}
}

func (h *MainWindow) GetOpenRepoMenu() cui.Menu {
	menu := h.ui.NewMenu("Open repo")
	menu.AddItems(h.OpenRepoMenuItems())
	return menu
}

func (h *MainWindow) GetStartMenu() cui.Menu {
	menu := h.ui.NewMenu("Open repo")
	menu.AddItems(h.OpenRepoMenuItems2())
	return menu
}

func (h *MainWindow) OpenRepo(folderPath string) {
	log.Infof("Opening %q ...", folderPath)
	workingFolder, err := git.WorkingFolderRoot(folderPath)
	if folderPath == "" || err != nil {
		log.Infof("No working folder %q", folderPath)
		openMenu := h.GetStartMenu()
		openMenu.Show(3, 1)
		return
	}
	log.Infof("Got repo %q ...", workingFolder)

	repoView := cviews.NewRepoView(h.ui, h.configService, h, workingFolder)
	repoView.Show()

	parent := filepath.Dir(workingFolder)
	h.configService.SetState(func(s *config.State) {
		s.RecentFolders = utils.RecentItems(s.RecentFolders, workingFolder, 10)
		s.RecentParentFolders = utils.RecentItems(s.RecentParentFolders, parent, 5)
	})
}

func (h *MainWindow) OpenRepoMenuItems() []cui.MenuItem {
	return []cui.MenuItem{
		h.RecentReposMenuItem(),
		h.getOpenMenuItem(),
	}
}

func (h *MainWindow) OpenRepoMenuItems2() []cui.MenuItem {
	items := h.getRecentMenuItems()
	if len(items) > 0 {
		items = append(items, cui.SeparatorMenuItem)
	}
	items = append(items, h.getOpenMenuItem())
	return items
}

func (h *MainWindow) RecentReposMenuItem() cui.MenuItem {
	return cui.MenuItem{Text: "Recent Repos", Title: "Recent Repos", SubItems: h.getRecentMenuItems()}
}

func (h *MainWindow) getRecentMenuItems() []cui.MenuItem {
	var items []cui.MenuItem
	for _, f := range h.configService.GetState().RecentFolders {
		path := f
		items = append(items, cui.MenuItem{Text: path, Action: func() {
			h.OpenRepo(path)
		}})
	}
	return items
}

func (h *MainWindow) getOpenMenuItem() cui.MenuItem {
	paths := h.configService.GetState().RecentParentFolders
	paths = append(paths, utils.GetVolumes()...)
	if len(paths) == 1 {
		return cui.MenuItem{Text: "Open", Title: paths[0], SubItemsFunc: func() []cui.MenuItem {
			return h.getFolderItems(paths[0], func(folder string) { h.OpenRepo(folder) })
		}}
	}

	var items []cui.MenuItem
	for _, p := range paths {
		path := p
		items = append(items,
			cui.MenuItem{Text: path, Title: path, SubItemsFunc: func() []cui.MenuItem {
				return h.getFolderItems(path, func(folder string) { h.OpenRepo(folder) })
			}})
	}
	return cui.MenuItem{Text: "Open Repo", SubItems: items}
}

func (h *MainWindow) getFolderItems(folder string, action func(f string)) []cui.MenuItem {
	files, err := ioutil.ReadDir(folder)
	if err != nil {
		// Folder not readable, might be e.g. access denied
		return nil
	}

	var items []cui.MenuItem

	for _, f := range files {
		if !f.IsDir() || f.Name() == "$RECYCLE.BIN" {
			continue
		}
		path := filepath.Join(folder, f.Name())
		items = append(items, cui.MenuItem{
			Text:   f.Name(),
			Title:  path,
			Action: func() { action(path) },
			SubItemsFunc: func() []cui.MenuItem {
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
