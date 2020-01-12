package repoview

import (
	"github.com/michael-reichenauer/gmc/common/config"
	"github.com/michael-reichenauer/gmc/repoview/viewmodel"
	"github.com/michael-reichenauer/gmc/utils/ui"
)

type showMode int

const (
	repo showMode = iota
	details
	detailsCurrent
)

type mainController interface {
	ToggleDetails()
}

type MainWindow struct {
	uiHandler     *ui.UI
	configService *config.Service
	model         *viewmodel.Service
	repoView      *RepoView
	detailsView   *DetailsView
	mode          showMode
}

func NewMainWindow(
	uiHandler *ui.UI,
	configService *config.Service,
	repoPath string) *MainWindow {
	m := viewmodel.NewModel(configService, repoPath)
	h := &MainWindow{
		uiHandler:     uiHandler,
		configService: configService,
		model:         m,
	}
	h.detailsView = newDetailsView(uiHandler, m)
	h.repoView = newRepoView(uiHandler, m, h.detailsView, h)
	return h
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
