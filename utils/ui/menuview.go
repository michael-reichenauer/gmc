package ui

import (
	"fmt"
	"github.com/jroimartin/gocui"
	"github.com/michael-reichenauer/gmc/utils"
	"strings"
)

const margin = 8

type menuView struct {
	View
	uiHandler       *UI
	currentViewName string
	items           []Item
	x               int
	y               int
	w               int
	h               int
	isMore          bool
	keyWidth        int
}

func newMenuView(uiHandler *UI, items []Item, x, y int) *menuView {
	h := &menuView{
		uiHandler: uiHandler,
		items:     items,
		x:         x,
		y:         y,
	}
	h.setSize()
	h.setPos()
	h.View = uiHandler.NewView(h.viewData)
	h.View.Properties().Name = "Menu"
	h.View.Properties().HasFrame = true
	return h
}

func (h *menuView) viewData(viewPort ViewPage) ViewData {
	var lines []string
	length := viewPort.FirstLine + viewPort.Height
	if length > len(h.items) {
		length = len(h.items)
	}
	for i := viewPort.FirstLine; i < length; i++ {
		lines = append(lines, h.toItemText(viewPort.Width, h.items[i]))
	}
	return ViewData{Lines: lines, FirstIndex: viewPort.FirstLine, Total: len(h.items)}
}

func (h *menuView) setPos() {
	windowWidth, windowHeight := h.uiHandler.WindowSize()
	if h.x+h.w > windowWidth-margin/2 {
		h.x = windowWidth - margin/2 - h.w
	}
	if h.y+h.h > windowHeight-margin/2 {
		h.y = windowHeight - margin/2 - h.h
	}
}

func (h *menuView) setSize() {
	windowWidth, windowHeight := h.uiHandler.WindowSize()
	h.w = h.maxWidth() + 4
	h.h = len(h.items) + 2
	if h.w < 10 {
		h.w = 10
	}
	if h.w > 30 {
		h.w = 30
	}
	if h.w > windowWidth-margin {
		h.w = windowWidth - margin
	}
	if h.h < 5 {
		h.h = 5
	}
	if h.h > 30 {
		h.h = 30
	}
	if h.h > windowHeight-margin {
		h.h = windowHeight - margin
	}
}

func (h *menuView) maxWidth() int {
	width := 0
	for _, item := range h.items {
		keyWidth := 0
		if len(item.Key) > 0 {
			keyWidth = len(item.Key) + 1
			if keyWidth > h.keyWidth {
				h.keyWidth = keyWidth
			}
		}
		moreWidth := 0
		if len(item.SubItems) > 0 {
			moreWidth = 1 + 1
			h.isMore = true
		}

		w := len(item.Text) + 1 + keyWidth + moreWidth
		if w > width {
			width = w
		}
	}
	return width
}

func (h *menuView) toItemText(width int, item Item) string {
	more := ""
	if h.isMore {
		more = "  "
	}
	if len(item.SubItems) > 0 {
		more = " >"
	}
	key := ""
	if h.keyWidth > 0 {
		key = strings.Repeat(" ", h.keyWidth+1)
	}
	if item.Key != "" {
		key = " " + utils.Text(item.Key, h.keyWidth)
	}

	return fmt.Sprintf("%s%s%s", utils.Text(item.Text, width-4), key, more)
}
func (h *menuView) show() {
	h.SetKey(gocui.KeyEsc, gocui.ModNone, h.onClose)
	h.SetKey(gocui.KeyEnter, gocui.ModNone, h.onEnter)
	h.SetKey(gocui.KeyArrowLeft, gocui.ModNone, h.onClose)
	h.SetKey(gocui.KeyArrowRight, gocui.ModNone, h.onSubItem)
	h.Show(Rect{X: h.x, Y: h.y, W: h.x + h.w, H: h.y + h.h})
	h.currentViewName = h.uiHandler.CurrentView()
	h.SetCurrentView()
	h.NotifyChanged()
}

func (h *menuView) onClose() {
	h.Close()
	h.uiHandler.SetCurrentView(h.currentViewName)
}

func (h *menuView) onEnter() {

}
func (h *menuView) onSubItem() {
	vp := h.ViewPage()
	y := h.y + (vp.CurrentLine - vp.FirstLine)
	mv := newMenuView(h.uiHandler, h.items, h.x+h.w, y)
	mv.show()
}
