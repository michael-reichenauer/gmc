package console

//
// import (
// 	"github.com/michael-reichenauer/gmc/common/config"
// 	"github.com/michael-reichenauer/gmc/utils"
// 	"github.com/michael-reichenauer/gmc/utils/cui"
// 	"github.com/michael-reichenauer/gmc/utils/git"
// 	"github.com/michael-reichenauer/gmc/utils/log"
// 	"io/ioutil"
// 	"path/filepath"
// 	"sort"
// 	"strings"
// )
//
// func (t *MainWindow) GetOpenRepoMenuItem() cui.MenuItem {
// 	return cui.MenuItem{Text: "Open repo", Title: "Open Repo", Key: "Ctrl-O",
// 		SubItems: t.OpenRepoMenuItems()}
// }
//
// func (t *MainWindow) GetOpenRepoMenu() cui.Menu {
// 	menu := t.ui.NewMenu("Open repo")
// 	menu.AddItems(t.OpenRepoMenuItems())
// 	return menu
// }
//
// func (t *MainWindow) GetStartMenu() cui.Menu {
// 	menu := t.ui.NewMenu("Open repo")
// 	menu.AddItems(t.OpenRepoMenuItems2())
// 	return menu
// }
//
// func (t *MainWindow) OpenRepo(folderPath string) {
// 	log.Infof("Opening %q ...", folderPath)
// 	workingFolder, err := git.WorkingDirRoot(folderPath)
// 	if folderPath == "" || err != nil {
// 		log.Infof("No working folder %q", folderPath)
// 		openMenu := t.GetStartMenu()
// 		openMenu.Show(3, 1)
// 		return
// 	}
// 	log.Infof("Got repo %q ...", workingFolder)
//
// 	repoView := NewRepoView(t.ui, t.configService, t, workingFolder)
// 	repoView.Show()
//
// 	parent := filepath.Dir(workingFolder)
// 	t.configService.SetState(func(s *config.State) {
// 		s.RecentFolders = utils.RecentItems(s.RecentFolders, workingFolder, 10)
// 		s.RecentParentFolders = utils.RecentItems(s.RecentParentFolders, parent, 5)
// 	})
// }
//
// func (t *MainWindow) OpenRepoMenuItems() []cui.MenuItem {
// 	return []cui.MenuItem{
// 		t.RecentReposMenuItem(),
// 		t.getOpenMenuItem(),
// 	}
// }
//
// func (t *MainWindow) OpenRepoMenuItems2() []cui.MenuItem {
// 	items := t.getRecentMenuItems()
// 	if len(items) > 0 {
// 		items = append(items, cui.SeparatorMenuItem)
// 	}
// 	items = append(items, t.getOpenMenuItem())
// 	return items
// }
//
// func (t *MainWindow) RecentReposMenuItem() cui.MenuItem {
// 	return cui.MenuItem{Text: "Recent Repos", Title: "Recent Repos", SubItems: t.getRecentMenuItems()}
// }
//
// func (t *MainWindow) getRecentRepoMenuItems(recentDirs []string) []cui.MenuItem {
// 	var items []cui.MenuItem
// 	for _, f := range recentDirs {
// 		path := f
// 		items = append(items, cui.MenuItem{Text: path, Action: func() {
// 			t.OpenRepo(path)
// 		}})
// 	}
// 	return items
// }
//
// func (t *MainWindow) getFolderItems(folder string, action func(f string)) []cui.MenuItem {
// 	files, err := ioutil.ReadDir(folder)
// 	if err != nil {
// 		// Folder not readable, might be e.g. access denied
// 		return nil
// 	}
//
// 	var items []cui.MenuItem
//
// 	for _, f := range files {
// 		if !f.IsDir() || f.Name() == "$RECYCLE.BIN" {
// 			continue
// 		}
// 		path := filepath.Join(folder, f.Name())
// 		items = append(items, cui.MenuItem{
// 			Text:   f.Name(),
// 			Title:  path,
// 			Action: func() { action(path) },
// 			SubItemsFunc: func() []cui.MenuItem {
// 				return t.getFolderItems(path, action)
// 			},
// 			ReuseBounds: true,
// 		})
// 	}
// 	sort.SliceStable(items, func(l, r int) bool {
// 		return -1 == strings.Compare(strings.ToLower(items[l].Text), strings.ToLower(items[r].Text))
// 	})
// 	return items
// }
