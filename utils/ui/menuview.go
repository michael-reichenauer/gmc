package ui

import (
	"github.com/jroimartin/gocui"
)

type menuView struct {
	View
	uiHandler       *UI
	currentViewName string
	items           []Item
	x               int
	y               int
}

func newMenuView(uiHandler *UI, items []Item, x, y int) *menuView {
	h := &menuView{
		uiHandler: uiHandler,
		items:     items,
		x:         x,
		y:         y,
	}
	h.View = uiHandler.NewView(h.viewData)
	h.View.Properties().Name = "Menu"
	h.View.Properties().HasFrame = true
	return h
}

func (h *menuView) viewData(viewPort ViewPage) ViewData {
	return ViewData{Lines: []string{"first line", "second"}}
}

func (h *menuView) show() {
	h.SetKey(gocui.KeyEsc, gocui.ModNone, h.onClose)
	h.SetKey(gocui.KeyEnter, gocui.ModNone, h.onEnter)
	h.SetKey(gocui.KeyArrowLeft, gocui.ModNone, h.onClose)
	h.SetKey(gocui.KeyArrowRight, gocui.ModNone, h.onEnter)
	h.Show(Rect{X: h.x, Y: h.y, W: h.x + 20, H: h.y + 10})
	h.currentViewName = h.uiHandler.CurrentView()
	h.SetCurrentView()
	h.NotifyChanged()
}

func (h *menuView) onClose() {
	h.Close()
	h.uiHandler.SetCurrentView(h.currentViewName)
}

func (h *menuView) onEnter() {
	mv := newMenuView(h.uiHandler, nil, h.x+3, h.y+3)
	mv.show()

}
