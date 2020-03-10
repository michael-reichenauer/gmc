package program

import (
	"fmt"
	"github.com/michael-reichenauer/gmc/common/config"
	"github.com/michael-reichenauer/gmc/repoview"
	"github.com/michael-reichenauer/gmc/repoview/viewmodel"
	"github.com/michael-reichenauer/gmc/utils"
	"github.com/michael-reichenauer/gmc/utils/git"
	"github.com/michael-reichenauer/gmc/utils/ui"
)

type showMode int

const (
	repo showMode = iota
	details
	detailsCurrent
)

type MainWindow struct {
	ui            *ui.UI
	configService *config.Service
	model         *viewmodel.Service
	repoView      *repoview.RepoView
	detailsView   *repoview.DetailsView
	diffView      *repoview.DiffView
	mode          showMode
}

func NewMainWindow(ui *ui.UI, configService *config.Service) *MainWindow {
	return &MainWindow{
		ui:            ui,
		configService: configService,
	}
}

func (h *MainWindow) NewMenu(title string) *ui.Menu {
	return h.ui.NewMenu(title)
}
func (h *MainWindow) Show() {
	workingFolder, err := h.getWorkingFolder()
	if err != nil {
		// Handle error
	}
	h.repoView = repoview.NewRepoView(h.ui, h.configService, h, workingFolder)
	h.repoView.Properties().HasFrame = false
	h.repoView.Show(ui.Rect{W: 1, H: 1})
	h.repoView.SetTop()
	h.repoView.SetCurrentView()

	h.OnResizeWindow()
}

func (h *MainWindow) ToggleShowDetails() {
	if h.mode == repo {
		h.mode = details
		h.detailsView = repoview.NewDetailsView(h.ui)
		h.detailsView.Show(ui.Rect{W: 1, H: 1})
		h.detailsView.SetTop()
	} else {
		h.mode = repo
		h.detailsView.Close()
		h.detailsView = nil
		h.repoView.SetTop()
	}
	h.OnResizeWindow()
}

func (h *MainWindow) ShowDiff(diffGetter repoview.DiffGetter, commitID string) {
	h.diffView = repoview.NewDiffView(h.ui, h, diffGetter, commitID)
	width, height := h.ui.WindowSize()
	h.diffView.Show(ui.Rect{X: 1, Y: 1, W: width - 2, H: height - 2})
	h.diffView.SetTop()
	h.diffView.SetCurrentView()
	h.OnResizeWindow()
}

func (h *MainWindow) HideDiff() {
	h.diffView.Close()
	h.diffView = nil
	h.repoView.SetCurrentView()
	h.repoView.SetTop()
	h.OnResizeWindow()
}

func (h *MainWindow) OnResizeWindow() {
	width, height := h.ui.WindowSize()
	if h.mode == repo {
		h.repoView.SetBounds(ui.Rect{X: 0, Y: 0, W: width, H: height})
		h.repoView.NotifyChanged()

		if h.diffView != nil {
			h.diffView.SetBounds(ui.Rect{X: 1, Y: 1, W: width - 2, H: height - 2})
			h.diffView.SetTop()
			h.diffView.NotifyChanged()
		}
	} else if h.mode == details {
		detailsHeight := 7
		h.repoView.SetBounds(ui.Rect{X: 0, Y: 0, W: width, H: height - detailsHeight - 1})
		h.repoView.NotifyChanged()
		h.detailsView.SetBounds(ui.Rect{X: 0, Y: height - detailsHeight - 1, W: width, H: detailsHeight + 1})
		h.detailsView.NotifyChanged()

		if h.diffView != nil {
			h.diffView.SetBounds(ui.Rect{X: 1, Y: 1, W: width - 2, H: height - 2})
			h.diffView.SetTop()
			h.diffView.NotifyChanged()
		}
	}
}

func (h *MainWindow) getWorkingFolder() (string, error) {
	folderPath := h.configService.FolderPath
	if folderPath == "" {
		// No specified repo path, use current dir
		folderPath = utils.CurrentDir()
	}
	path, err := git.WorkingFolderRoot(folderPath)
	if err != nil {
		return "", err
	}
	return path, nil
}

func (h *MainWindow) MainMenuItem() ui.MenuItem {
	var subItems []ui.MenuItem
	subItems = append(subItems, h.OpenRepoMenuItems()...)
	subItems = append(subItems, ui.MenuItem{Text: "About", Action: h.showAbout})
	menuItem := ui.MenuItem{
		Text:     "Main Menu",
		Title:    "Main Menu",
		SubItems: subItems,
	}

	return menuItem
}

func (h *MainWindow) showAbout() {
	msgBox := ui.NewMessageBox(h.ui, fmt.Sprintf("gmc %s", h.configService.ProgramVersion), "About")
	msgBox.Show()
}
