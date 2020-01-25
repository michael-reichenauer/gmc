package program

import (
	"fmt"
	"github.com/michael-reichenauer/gmc/common/config"
	"github.com/michael-reichenauer/gmc/repoview"
	"github.com/michael-reichenauer/gmc/repoview/viewmodel"
	"github.com/michael-reichenauer/gmc/utils"
	"github.com/michael-reichenauer/gmc/utils/gitlib"
	"github.com/michael-reichenauer/gmc/utils/log"
	"github.com/michael-reichenauer/gmc/utils/ui"
)

type showMode int

const (
	repo showMode = iota
	details
	detailsCurrent
)

type MainWindow struct {
	uiHandler     *ui.UI
	configService *config.Service
	model         *viewmodel.Service
	repoView      *repoview.RepoView
	detailsView   *repoview.DetailsView
	mode          showMode
}

func NewMainWindow(uiHandler *ui.UI, configService *config.Service) *MainWindow {
	h := &MainWindow{
		uiHandler:     uiHandler,
		configService: configService,
	}
	workingFolder := h.getWorkingFolder()
	vm := viewmodel.NewModel(configService, workingFolder)
	h.detailsView = repoview.NewDetailsView(uiHandler, vm)
	h.repoView = repoview.NewRepoView(uiHandler, vm, h.detailsView, h)
	return h
}

func (h *MainWindow) ShowAbout() {
	msgBox := ui.NewMessageBox(h.uiHandler, fmt.Sprintf("gmc %s", h.configService.ProgramVersion), "About")
	msgBox.Show()
}

func (h *MainWindow) Show() {
	r := ui.Rect{W: 1, H: 1}
	h.repoView.Properties().HasFrame = false
	h.detailsView.Show(r)
	h.repoView.Show(r)
	h.repoView.SetCurrentView()

	h.OnResizeWindow()
}

func (h *MainWindow) ToggleDetails() {
	if h.mode == repo {
		h.mode = details
	} else {
		h.mode = repo
	}
	h.OnResizeWindow()
}

func (h *MainWindow) OnResizeWindow() {
	width, height := h.uiHandler.WindowSize()
	if h.mode == repo {
		h.repoView.SetBounds(ui.Rect{X: 0, Y: 0, W: width, H: height})
		h.detailsView.SetBounds(ui.Rect{X: -1, Y: -1, W: 1, H: 1})
	} else if h.mode == details {
		detailsHeight := 7
		h.repoView.SetBounds(ui.Rect{X: 0, Y: 0, W: width, H: height - detailsHeight - 1})
		h.detailsView.SetBounds(ui.Rect{X: 0, Y: height - detailsHeight - 1, W: width, H: detailsHeight + 1})
	}
	h.repoView.NotifyChanged()
	h.detailsView.NotifyChanged()
}

func (h *MainWindow) getWorkingFolder() string {
	folderPath := h.configService.FolderPath
	if folderPath == "" {
		// No specified repo path, use current dir
		folderPath = utils.CurrentDir()
	}
	path, err := gitlib.WorkingFolderRoot(folderPath)
	if err != nil {
		panic(log.Fatal(err))
	}
	return path
}
