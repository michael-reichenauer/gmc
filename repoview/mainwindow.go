package repoview

import (
	"github.com/michael-reichenauer/gmc/repoview/model"
	"github.com/michael-reichenauer/gmc/utils/ui"
)

type MainWindow struct {
	uiHandler   *ui.UI
	model       *model.Model
	repoView    *RepoView
	detailsView *DetailsView
}

func NewMainWindow(uiHandler *ui.UI, repoPath string) *MainWindow {
	m := model.NewModel(repoPath)
	detailsView := newDetailsView(uiHandler, m)
	repoView := newRepoView(uiHandler, m, detailsView)
	return &MainWindow{
		uiHandler:   uiHandler,
		model:       m,
		repoView:    repoView,
		detailsView: detailsView,
	}
}

func (h *MainWindow) Show() {
	r := ui.Rect{0, 0, 1, 1}
	h.repoView.Properties().HasFrame = true
	h.repoView.Show(r)
	h.detailsView.Show(r)
	h.repoView.SetCurrentView()

	h.OnResizeWindow()
}

func (h *MainWindow) OnResizeWindow() {
	width, height := h.uiHandler.WindowSize()
	h.repoView.SetBounds(ui.Rect{X: 0, Y: 0, W: width, H: height - 5})
	h.detailsView.SetBounds(ui.Rect{X: 0, Y: height - 4, W: width, H: 4})

	h.repoView.NotifyChanged()
	h.detailsView.NotifyChanged()
}
