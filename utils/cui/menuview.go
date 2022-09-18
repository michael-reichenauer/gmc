package cui

import (
	"fmt"
	"strings"

	"github.com/jroimartin/gocui"
	"github.com/michael-reichenauer/gmc/utils"
	"github.com/michael-reichenauer/gmc/utils/log"
)

const margin = 8

type menuView struct {
	View
	ui           *ui
	parent       *menuView
	title        string
	items        []MenuItem
	bounds       Rect
	moreWidth    int
	keyWidth     int
	marginsWidth int
}

func newMenuView(ui *ui, title string, parent *menuView) *menuView {
	h := &menuView{ui: ui, parent: parent, title: title}
	h.View = ui.NewViewFromPageFunc(h.viewData)
	h.View.Properties().Name = "Menu"
	h.View.Properties().HasFrame = true
	h.View.Properties().Title = title
	h.View.Properties().OnMouseOutside = h.onClose
	h.View.Properties().OnMouseLeft = h.onMouseLeft
	h.View.Properties().HideHorizontalScrollbar = true
	h.View.Properties().IsMoveUpDownWrap = true
	return h
}

func (t *menuView) addItems(items []MenuItem) {
	t.items = append(t.items, items...)
}

func (t *menuView) show(bounds Rect) {
	t.bounds = t.getBounds(t.items, bounds)
	t.SetKey(gocui.KeyEsc, t.onClose)
	t.SetKey(gocui.KeyEnter, t.onEnter)
	t.SetKey(gocui.KeyArrowLeft, t.onClose)
	t.SetKey(gocui.KeyArrowRight, t.onSubItem)
	t.Show(Bounds(Rect{X: t.bounds.X, Y: t.bounds.Y, W: t.bounds.W, H: t.bounds.H}))
	t.SetCurrentView()
	t.NotifyChanged()
}

func (t *menuView) viewData(viewPort ViewPage) ViewText {
	var lines []string
	length := viewPort.FirstLine + viewPort.Height
	if length > len(t.items) {
		length = len(t.items)
	}

	for i := viewPort.FirstLine; i < length; i++ {
		line := t.toItemText(viewPort.Width, t.items[i])
		lines = append(lines, line)
	}
	return ViewText{Lines: lines, Total: len(t.items)}
}

func (t *menuView) getBounds(items []MenuItem, bounds Rect) Rect {
	width, height := t.getSize(items)
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

	x2, y2 := t.getPos(bounds.X, bounds.Y, width, height)
	return Rect{X: x2, Y: y2, W: width, H: height}
}

func (t *menuView) getPos(x1, y1, width, height int) (x int, y int) {
	windowWidth, windowHeight := t.ui.WindowSize()
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

func (t *menuView) getSize(items []MenuItem) (width, height int) {
	windowWidth, windowHeight := t.ui.WindowSize()

	width, t.keyWidth, t.moreWidth, t.marginsWidth = t.maxWidth(items)
	if width < 10 {
		width = 10
	}
	if width > 100 {
		width = 100
	}
	if width > windowWidth-margin {
		width = windowWidth - margin
	}

	height = len(t.items)
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

func (t *menuView) maxWidth(items []MenuItem) (maxWidth, maxKeyWidth, maxMoreWidth, marginsWidth int) {
	maxTextWidth := t.maxTextWidth(items)
	maxKeyWidth = t.maxKeyWidth(items)
	maxMoreWidth = t.maxMoreWidth(items)

	marginsWidth = 0
	if maxKeyWidth > 0 {
		marginsWidth = marginsWidth + 3
	}
	if maxMoreWidth > 0 {
		marginsWidth++
	}
	maxWidth = maxTextWidth + maxKeyWidth + maxMoreWidth + marginsWidth
	titleWidth := len(t.title) + 3
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
		if len(item.SubItems) > 0 || item.SubItemsFunc != nil {
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

func (t *menuView) toItemText(width int, item MenuItem) string {
	key := ""
	if t.keyWidth > 0 {
		if item.Key != "" {
			key = "   " + utils.Text(item.Key, t.keyWidth)
		} else {
			key = "  " + strings.Repeat(" ", t.keyWidth+1)
		}
	}

	more := ""
	if t.moreWidth > 0 {
		if len(item.SubItems) > 0 || item.SubItemsFunc != nil {
			more = " ►"
		} else {
			more = "  "
		}
	}

	extraWidth := t.marginsWidth + t.keyWidth + t.moreWidth
	text := utils.Text(item.Text, width-extraWidth)
	if item.isSeparator {
		text = strings.Repeat("─", width-extraWidth)
	}
	return fmt.Sprintf("%s%s%s", text, key, more)
}

func (t *menuView) onClose() {
	log.Debugf("On close")
	t.Close()
}

func (t *menuView) closeAll() {
	log.Debugf("Close all from")
	if t.parent == nil {
		t.onClose()
		return
	}
	t.Close()
	t.parent.closeAll()
}

func (t *menuView) onEnter() {
	vp := t.ViewPage()
	log.Debugf("enter %d", vp.CurrentLine)
	t.action(vp.FirstLine, vp.CurrentLine, false)
}

func (t *menuView) onSubItem() {
	vp := t.ViewPage()
	t.subItem(vp.FirstLine, vp.CurrentLine)
}

func (t *menuView) onMouseLeft(x int, y int) {
	vp := t.ViewPage()
	log.Debugf("mouse %d", vp.FirstLine+y)
	isMoreClicked := vp.Width-x < 2
	t.action(vp.FirstLine, vp.FirstLine+y, isMoreClicked)
}

func (t *menuView) action(firstLine, index int, isMoreClicked bool) {
	item := t.items[index]
	log.Debugf("action item %d %q, %v", index, item.Text, item.Action == nil)

	hasSubActions := len(item.SubItems) != 0 || item.SubItemsFunc != nil
	if hasSubActions && (item.Action == nil || isMoreClicked) {
		t.subItem(firstLine, index)
		return
	}

	if item.Action == nil {
		return
	}
	t.closeAll()
	item.Action()
}

func (t *menuView) subItem(firstLine, index int) {
	if index >= len(t.items) {
		return
	}
	item := t.items[index]
	log.Debugf("sub item %d %q", index, item.Text)

	if item.SubItemsFunc != nil {
		t.showSubItemsPlaceholderMenu(firstLine, index, item, item.SubItemsFunc)
		return
	}
	if len(item.SubItems) == 0 {
		return
	}

	t.showSubItemsMenu(firstLine, index, item, item.SubItems)
}

func (t *menuView) showSubItemsMenu(firstLine, index int, item MenuItem, subItems []MenuItem) *menuView {
	var subBonds Rect
	subBonds.X = t.bounds.X + t.bounds.W
	windowWidth, _ := t.ui.WindowSize()
	maxSubWidth, _, _, _ := t.maxWidth(subItems)
	if subBonds.X+maxSubWidth > windowWidth {
		subBonds.X = t.bounds.X - maxSubWidth
	}

	subBonds.Y = t.bounds.Y + (index - firstLine)

	mv := newMenuView(t.ui, item.Title, t)
	mv.addItems(subItems)
	if item.ReuseBounds {
		subBonds = t.bounds
	}
	mv.show(subBonds)
	return mv
}

func (t *menuView) showSubItemsPlaceholderMenu(firstLine, index int, item MenuItem, subItemsFunc func() []MenuItem) {
	subItems := []MenuItem{{Text: " "}, {Text: "Retrieving ..."}, {Text: " "}}
	mv := t.showSubItemsMenu(firstLine, index, item, subItems)
	p := t.ui.ShowProgress("Getting %s ...", item.Text)

	go func() {
		items := subItemsFunc()
		t.ui.Post(func() {
			p.Close()
			mv.Close()
			t.showSubItemsMenu(firstLine, index, item, items)
		})
	}()
}
