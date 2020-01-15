package ui

import (
	"fmt"
	"github.com/jroimartin/gocui"
	"github.com/michael-reichenauer/gmc/utils"
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
	var lines []string
	for i := 0; i < len(h.items); i++ {
		lines = append(lines, h.toItemText(viewPort.Width, h.items[i]))
	}
	return ViewData{Lines: lines}
}

func (h *menuView) toItemText(width int, item Item) string {
	more := " "
	if len(item.SubItems) > 0 {
		more = ">"
	}
	return fmt.Sprintf("%s %s %s", utils.Text(item.Text, width-5), item.Key, more)
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
	mv := newMenuView(h.uiHandler, h.items, h.x+3, h.y+3)
	mv.show()

}
