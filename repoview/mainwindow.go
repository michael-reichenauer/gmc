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
	r := ui.Rect{0, 0, 1, 1}
	h.repoView.Properties().HasFrame = true
	h.statusView.Show(r)
	h.repoView.Show(r)
	h.repoView.SetCurrentView()

	h.OnResizeWindow()
}

func (h *MainWindow) OnResizeWindow() {
	width, height := h.uiHandler.WindowSize()
	log.Infof("Resize %d %d", width, height)
	h.statusView.SetBounds(ui.Rect{X: 0, Y: 0, W: width, H: 1})
	h.repoView.SetBounds(ui.Rect{X: 0, Y: 2, W: width, H: height})

	h.statusView.NotifyChanged()
	h.repoView.NotifyChanged()
}
