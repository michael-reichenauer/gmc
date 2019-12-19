package repoview

import (
	"github.com/michael-reichenauer/gmc/utils/log"
	"github.com/michael-reichenauer/gmc/utils/ui"
)

type MainWindow struct {
	uiHandler  *ui.Handler
	statusView *StatusViewHandler
	repoView   *RepoView
}

func NewMainWindow(uiHandler *ui.Handler, repoPath string) *MainWindow {
	return &MainWindow{
		uiHandler:  uiHandler,
		statusView: NewStatusView(uiHandler),
		repoView:   NewRepoView(uiHandler, repoPath),
	}
}

func (h *MainWindow) Show() {
	width, height := h.uiHandler.WindowSize()
	h.statusView.Show(ui.Rect{X: 0, Y: 0, W: width, H: 1})
	h.repoView.Show(ui.Rect{X: 0, Y: 1, W: width, H: height - 1})
	h.repoView.SetCurrentView()
}

func (h *MainWindow) OnResizeWindow() {
	width, height := h.uiHandler.WindowSize()
	log.Infof("Resize %d %d", width, height)
	h.statusView.SetBounds(ui.Rect{X: 0, Y: 0, W: width, H: 1})
	h.repoView.SetBounds(ui.Rect{X: 0, Y: 1, W: width, H: height - 1})
	h.statusView.NotifyChanged()
	h.repoView.NotifyChanged()
}
