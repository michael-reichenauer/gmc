package console

import (
	"fmt"
	"github.com/michael-reichenauer/gmc/api"
	"github.com/michael-reichenauer/gmc/client"
	"github.com/michael-reichenauer/gmc/utils/cui"
	"github.com/michael-reichenauer/gmc/utils/log"
	"github.com/michael-reichenauer/gmc/utils/rpc"
	"path"
	"strings"
	"time"
)

type MainWindow struct {
	ui        cui.UI
	api       api.Api
	rpcClient *rpc.Client
}

func NewMainWindow(ui cui.UI) *MainWindow {
	return &MainWindow{ui: ui}
}

func (t *MainWindow) Show(serverUri, path string) {
	progress := t.ui.ShowProgress("Connecting")
	go func() {
		// Create rpc client and create service client
		rpcClient := rpc.NewClient()
		rpcClient.IsLogCalls = true
		rpcClient.Latency = 500 * time.Millisecond
		rpcClient.OnConnectionError = func(err error) {
			t.ui.PostOnUIThread(func() {
				msgBox := t.ui.MessageBox("Error !", cui.Red(fmt.Sprintf("Connection to server failed:\n%v", err)))
				msgBox.OnOK = func() { t.ui.PostOnUIThread(func() { t.Show(serverUri, path) }) }
				msgBox.Show()
			})
		}
		time.AfterFunc(20*time.Second, func() {
			//	rpcClient.Interrupt()
		})
		err := rpcClient.Connect(serverUri)
		api := client.NewClient(rpcClient.ServiceClient(""))

		t.ui.PostOnUIThread(func() {
			if err != nil {
				t.ui.ShowErrorMessageBox("Failed to connect:\n%v", err)
				return
			}
			t.api = api
			t.rpcClient = rpcClient
			progress.Close()

			t.showRepo(path)
		})
	}()
}

func (t *MainWindow) showRepo(path string) {
	progress := t.ui.ShowProgress("Opening repo:\n%s", path)
	go func() {
		err := t.api.OpenRepo(path, api.NilRsp)
		t.ui.PostOnUIThread(func() {
			if err != nil {
				log.Warnf("Failed to open %q, %v", path, err)
				if path != "" {
					t.ui.ShowErrorMessageBox("Failed to show repo for:\n%s\nError: %v", path, err)
				}
				t.showOpenRepoMenu()
				return
			}

			progress.Close()
			repoView := NewRepoView(t.ui, t.api)
			repoView.Show()
		})
	}()
}

func (t *MainWindow) Close() {
	t.rpcClient.Close()
}

func (t *MainWindow) showOpenRepoMenu() {
	menu := t.ui.NewMenu("Open repo")

	var recentDirs []string
	err := t.api.GetRecentWorkingDirs(api.NilArg, &recentDirs)
	if err != nil {
		t.ui.ShowErrorMessageBox("Failed to get recent dirs,\nError: %v", err)
	}
	if len(recentDirs) > 0 {
		menu.Add(t.getRecentRepoMenuItems(recentDirs)...)
		menu.Add(cui.SeparatorMenuItem)
	}

	var paths []string
	err = t.api.GetSubDirs("", &paths)
	if err != nil {
		t.ui.ShowErrorMessageBox("Failed to list of folders to open,\nError: %v", err)
		return
	}

	openItemsFunc := func() []cui.MenuItem {
		return t.getDirItems(paths, func(path string) { t.showRepo(path) })
	}

	menu.Add(cui.MenuItem{Text: "Open Repo", SubItemsFunc: openItemsFunc})

	menu.Show(3, 1)
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
			SubItemsFunc: func() []cui.MenuItem {
				var dirs []string
				err := t.api.GetSubDirs(path, &dirs)
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
