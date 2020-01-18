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
	parent          *menuView
	currentViewName string
	items           []MenuItem
	bounds          Rect
	moreWidth       int
	keyWidth        int
	marginsWidth    int
}

func newMenuView(uiHandler *UI, parent *menuView, items []MenuItem, x, y int) *menuView {
	h := &menuView{
		uiHandler: uiHandler,
		parent:    parent,
		items:     items,
	}

	h.View = uiHandler.NewView(h.viewData)
	h.View.Properties().Name = "Menu"
	h.View.Properties().HasFrame = true
	h.bounds = h.getBounds(items, x, y)
	return h
}

func (h *menuView) viewData(viewPort ViewPage) ViewData {
	var lines []string
	length := viewPort.FirstLine + viewPort.Height
	if length > len(h.items) {
		length = len(h.items)
	}

	for i := viewPort.FirstLine; i < length; i++ {
		line := h.toItemText(viewPort.Width, h.items[i])
		lines = append(lines, line)
	}
	return ViewData{Lines: lines, FirstIndex: viewPort.FirstLine, Total: len(h.items)}
}

func (h *menuView) getBounds(items []MenuItem, x, y int) Rect {
	width, height := h.getSize(items)
	x2, y2 := h.getPos(x, y, width, height)
	return Rect{X: x2, Y: y2, W: width, H: height}
}

func (h *menuView) getPos(x1, y1, width, height int) (x int, y int) {
	windowWidth, windowHeight := h.uiHandler.WindowSize()
	if x1 < 3 {
		x1 = 3
	}
	if y1 < 3 {
		y1 = 3
	}

	if x1+width > windowWidth-2 {
		x1 = windowWidth - 2 - width
	}
	if y1+height > windowHeight-1 {
		y1 = windowHeight - 1 - height
	}
	return x1, y1
}

func (h *menuView) getSize(items []MenuItem) (width, height int) {
	windowWidth, windowHeight := h.uiHandler.WindowSize()

	width, h.keyWidth, h.moreWidth, h.marginsWidth = h.maxWidth(items)
	if width < 10 {
		width = 10
	}
	if width > 100 {
		width = 100
	}
	if width > windowWidth-margin {
		width = windowWidth - margin
	}

	height = len(h.items)
	if height < 0 {
		height = 0
	}
	if height > 30 {
		height = 30
	}
	if height > windowHeight-2 {
		height = windowHeight - 2
	}
	return width, height
}

func (h *menuView) maxWidth(items []MenuItem) (maxWidth, maxKeyWidth, maxMoreWidth, marginsWidth int) {
	maxTextWidth := h.maxTextWidth(items)
	maxKeyWidth = h.maxKeyWidth(items)
	maxMoreWidth = h.maxMoreWidth(items)

	marginsWidth = 0
	if maxKeyWidth > 0 {
		marginsWidth = marginsWidth + 3
	}
	if maxMoreWidth > 0 {
		marginsWidth++
	}
	maxWidth = maxTextWidth + maxKeyWidth + maxMoreWidth + marginsWidth
	return
}

func (*menuView) maxKeyWidth(items []MenuItem) int {
	maxKeyWidth := 0
	for _, item := range items {
		keyWidth := 0
		if len(item.Key) > 0 {
			keyWidth = len(item.Key)
			if keyWidth > maxKeyWidth {
				maxKeyWidth = keyWidth
			}
		}
	}
	return maxKeyWidth
}

func (*menuView) maxMoreWidth(items []MenuItem) int {
	maxMoreWidth := 0
	for _, item := range items {
		moreWidth := 0
		if len(item.SubItems) > 0 {
			moreWidth = 1
			if moreWidth > maxMoreWidth {
				maxMoreWidth = moreWidth
			}
		}
	}
	return maxMoreWidth
}

func (*menuView) maxTextWidth(items []MenuItem) int {
	maxTextWidth := 0
	for _, item := range items {
		textWidth := len(item.Text)
		if textWidth > maxTextWidth {
			maxTextWidth = textWidth
		}
	}
	return maxTextWidth + 2
}

func (h *menuView) toItemText(width int, item MenuItem) string {
	key := ""
	if h.keyWidth > 0 {
		if item.Key != "" {
			key = "   " + utils.Text(item.Key, h.keyWidth)
		} else {
			key = strings.Repeat(" ", h.keyWidth+1)
		}
	}

	more := ""
	if h.moreWidth > 0 {
		if len(item.SubItems) > 0 {
			more = " ►"
		} else {
			more = "  "
		}
	}

	extraWidth := h.marginsWidth + h.keyWidth + h.moreWidth
	return fmt.Sprintf("%s%s%s", utils.Text(item.Text, width-extraWidth), key, more)
}

func (h *menuView) show() {
	h.currentViewName = h.uiHandler.CurrentView()
	h.SetKey(gocui.KeyEsc, gocui.ModNone, h.onClose)
	h.SetKey(gocui.KeyEnter, gocui.ModNone, h.onEnter)
	h.SetKey(gocui.KeyArrowLeft, gocui.ModNone, h.onClose)
	h.SetKey(gocui.KeyArrowRight, gocui.ModNone, h.onSubItem)
	h.Show(Rect{X: h.bounds.X, Y: h.bounds.Y, W: h.bounds.X + h.bounds.W, H: h.bounds.Y + h.bounds.H})
	h.SetCurrentView()
	h.NotifyChanged()
}

func (h *menuView) onClose() {
	h.Close()
	h.uiHandler.SetCurrentView(h.currentViewName)
}

func (h *menuView) closeAll() {
	if h.parent == nil {
		h.onClose()
		return
	}
	h.Close()
	h.parent.closeAll()
}

func (h *menuView) onEnter() {
	vp := h.ViewPage()
	h.closeAll()

	item := h.items[vp.CurrentLine]
	if item.Action == nil {
		return
	}
	item.Action()
}

func (h *menuView) onSubItem() {
	vp := h.ViewPage()
	item := h.items[vp.CurrentLine]
	if len(item.SubItems) == 0 {
		return
	}
	items := item.SubItems
	x := h.bounds.X + h.bounds.W
	windowWidth, _ := h.uiHandler.WindowSize()
	maxSubWidth, _, _, _ := h.maxWidth(items)
	if x+maxSubWidth > windowWidth {
		x = h.bounds.X - maxSubWidth
	}

	y := h.bounds.Y + (vp.CurrentLine - vp.FirstLine)
	mv := newMenuView(h.uiHandler, h, items, x, y)
	mv.show()
}
