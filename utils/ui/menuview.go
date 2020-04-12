package ui

import (
	"fmt"
	"github.com/jroimartin/gocui"
	"github.com/michael-reichenauer/gmc/utils"
	"github.com/michael-reichenauer/gmc/utils/log"
	"strings"
)

const margin = 8

type menuView struct {
	View
	ui           *UI
	parent       *menuView
	title        string
	items        []MenuItem
	bounds       Rect
	moreWidth    int
	keyWidth     int
	marginsWidth int
}

func newMenuView(ui *UI, title string, parent *menuView) *menuView {
	h := &menuView{ui: ui, parent: parent, title: title}
	h.View = ui.NewViewFromPageFunc(h.viewData)
	h.View.Properties().Name = "Menu"
	h.View.Properties().HasFrame = true
	h.View.Properties().Title = title
	h.View.Properties().OnMouseOutside = h.onClose
	h.View.Properties().OnMouseLeft = h.onMouseLeft
	h.View.Properties().HideHorizontalScrollbar = true
	return h
}

func (h *menuView) addItems(items []MenuItem) {
	h.items = append(h.items, items...)
}

func (h *menuView) show(bounds Rect) {
	h.bounds = h.getBounds(h.items, bounds)
	h.SetKey(gocui.KeyEsc, h.onClose)
	h.SetKey(gocui.KeyEnter, h.onEnter)
	h.SetKey(gocui.KeyArrowLeft, h.onClose)
	h.SetKey(gocui.KeyArrowRight, h.onSubItem)
	h.Show(Bounds(Rect{X: h.bounds.X, Y: h.bounds.Y, W: h.bounds.W, H: h.bounds.H}))
	h.SetCurrentView()
	h.NotifyChanged()
}

func (h *menuView) viewData(viewPort ViewPage) ViewText {
	var lines []string
	length := viewPort.FirstLine + viewPort.Height
	if length > len(h.items) {
		length = len(h.items)
	}

	for i := viewPort.FirstLine; i < length; i++ {
		line := h.toItemText(viewPort.Width, h.items[i])
		lines = append(lines, line)
	}
	return ViewText{Lines: lines, Total: len(h.items)}
}

func (h *menuView) getBounds(items []MenuItem, bounds Rect) Rect {
	width, height := h.getSize(items)
	if bounds.W != 0 {
		if width < bounds.W {
			width = bounds.W
		}
	}
	if bounds.H != 0 {
		if height < bounds.H {
			height = bounds.H
		}
	}

	x2, y2 := h.getPos(bounds.X, bounds.Y, width, height)
	return Rect{X: x2, Y: y2, W: width, H: height}
}

func (h *menuView) getPos(x1, y1, width, height int) (x int, y int) {
	windowWidth, windowHeight := h.ui.WindowSize()
	if x1 < 3 {
		x1 = 1
	}
	if y1 < 3 {
		y1 = 1
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
	windowWidth, windowHeight := h.ui.WindowSize()

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
	titleWidth := len(h.title) + 3
	if maxWidth < titleWidth {
		maxWidth = titleWidth
	}
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
		if len(item.SubItems) > 0 || item.SubItemsFunc != nil && len(item.SubItemsFunc()) > 0 {
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
			key = "  " + strings.Repeat(" ", h.keyWidth+1)
		}
	}

	more := ""
	if h.moreWidth > 0 {
		if len(item.SubItems) > 0 || item.SubItemsFunc != nil && len(item.SubItemsFunc()) > 0 {
			more = " ►"
		} else {
			more = "  "
		}
	}

	extraWidth := h.marginsWidth + h.keyWidth + h.moreWidth
	text := utils.Text(item.Text, width-extraWidth)
	if item.isSeparator {
		text = strings.Repeat("─", width-extraWidth)
	}
	return fmt.Sprintf("%s%s%s", text, key, more)
}

func (h *menuView) onClose() {
	log.Infof("On close")
	h.Close()
}

func (h *menuView) closeAll() {
	log.Infof("Close all from")
	if h.parent == nil {
		h.onClose()
		return
	}
	h.Close()
	h.parent.closeAll()
}

func (h *menuView) onEnter() {
	vp := h.ViewPage()
	log.Infof("enter %d", vp.CurrentLine)
	h.action(vp.FirstLine, vp.CurrentLine, false)
}

func (h *menuView) onSubItem() {
	vp := h.ViewPage()
	h.subItem(vp.FirstLine, vp.CurrentLine)
}

func (h *menuView) onMouseLeft(x int, y int) {
	vp := h.ViewPage()
	log.Infof("mouse %d", vp.FirstLine+y)
	isMoreClicked := vp.Width-x < 2
	h.action(vp.FirstLine, vp.FirstLine+y, isMoreClicked)
}

func (h *menuView) action(firstLine, index int, isMoreClicked bool) {
	log.Infof("action %d", index)
	item := h.items[index]
	log.Infof("action item %d %q, %v", index, item.Text, item.Action == nil)

	var subItems []MenuItem
	if item.SubItemsFunc != nil {
		subItems = item.SubItemsFunc()
	} else {
		subItems = item.SubItems
	}
	if len(subItems) != 0 && isMoreClicked {
		h.subItem(firstLine, index)
		return
	}

	if item.Action == nil {
		return
	}
	h.closeAll()
	item.Action()
}

func (h *menuView) subItem(firstLine, index int) {
	if index >= len(h.items) {
		return
	}
	item := h.items[index]
	log.Infof("sub item %d %q", index, item.Text)

	var subItems []MenuItem
	if item.SubItemsFunc != nil {
		subItems = item.SubItemsFunc()
	} else {
		subItems = item.SubItems
	}
	if len(subItems) == 0 {
		return
	}
	var subBonds Rect
	subBonds.X = h.bounds.X + h.bounds.W
	windowWidth, _ := h.ui.WindowSize()
	maxSubWidth, _, _, _ := h.maxWidth(subItems)
	if subBonds.X+maxSubWidth > windowWidth {
		subBonds.X = h.bounds.X - maxSubWidth
	}

	subBonds.Y = h.bounds.Y + (index - firstLine)
	mv := newMenuView(h.ui, item.Title, h)
	mv.addItems(subItems)
	if item.ReuseBounds {
		subBonds = h.bounds
	}
	mv.show(subBonds)
}
