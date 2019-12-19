package repoview

import (
	"github.com/michael-reichenauer/gmc/repoview/model"
	"github.com/michael-reichenauer/gmc/utils/log"
	"github.com/michael-reichenauer/gmc/utils/ui"
)

type MainWindow struct {
	uiHandler  *ui.UI
	model      *model.Model
	statusView *StatusViewHandler
	repoView   *RepoView
}

func NewMainWindow(uiHandler *ui.UI, repoPath string) *MainWindow {
	m := model.NewModel(repoPath)
	return &MainWindow{
		uiHandler:  uiHandler,
		model:      m,
		statusView: newStatusView(uiHandler, m),
		repoView:   newRepoView(uiHandler, m),
	}
}

func (h *MainWindow) Show() {
	width, height := h.uiHandler.WindowSize()
	h.statusView.Show(ui.Rect{X: 0, Y: 0, W: width, H: 1})
	h.repoView.Show(ui.Rect{X: 0, Y: 1, W: width, H: height - 1})
	h.repoView.SetCurrentView()
	h.statusView.NotifyChanged()
	h.repoView.NotifyChanged()
}

func (h *MainWindow) OnResizeWindow() {
	width, height := h.uiHandler.WindowSize()
	log.Infof("Resize %d %d", width, height)
	h.statusView.SetBounds(ui.Rect{X: 0, Y: 0, W: width, H: 1})
	h.repoView.SetBounds(ui.Rect{X: 0, Y: 1, W: width, H: height - 1})

	h.statusView.NotifyChanged()
	h.repoView.NotifyChanged()
}
