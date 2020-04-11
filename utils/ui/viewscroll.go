package ui

import (
	"github.com/jroimartin/gocui"
	"github.com/michael-reichenauer/gmc/utils/log"
	"math"
	"strings"
)

const (
	scrollBarVerticalHandle   = '┃' // The vertical scrollbar handle (right)
	scrollBarHorizontalHandle = '━' // The horizontal scrollbar handle (down)
)

func (h *view) createVerticalScrollView() *gocui.View {
	view := h.ui.createView()
	view.Frame = false
	h.ui.setKey(view, gocui.MouseLeft, h.onVerticalScrollMouseLeftClick)
	return view
}

func (h *view) createHorizontalScrollView() *gocui.View {
	view := h.ui.createView()
	view.Frame = false
	h.ui.setKey(view, gocui.MouseLeft, h.onHorizontalScrollMouseLeftClick)
	return view
}

func (h *view) onVerticalScrollMouseLeftClick() {
	_, cy := h.vertScrlView.Cursor()

	currentView := h.ui.currentView()
	if h != currentView && currentView.properties.OnMouseOutside != nil {
		// Mouse down, but this is not the current view, inform the current view
		currentView.properties.OnMouseOutside()
		return
	}

	if h.hasVerticalScrollbar() {
		// Mouse down in vertical scrollbar, set scrollbar to that position
		if h.isScrollHorizontal {
			// Vertical scrollbar not active, let activate first
			h.toggleScrollDirection()
			return
		}
		h.setVerticalScroll(cy)
		return
	}
}
func (h *view) onHorizontalScrollMouseLeftClick() {
	cx, _ := h.horzScrlView.Cursor()

	currentView := h.ui.currentView()
	if h != currentView && currentView.properties.OnMouseOutside != nil {
		// Mouse down, but this is not the current view, inform the current view
		currentView.properties.OnMouseOutside()
		return
	}

	if h.hasHorizontalScrollbar() {
		// Mouse down in horizontal scrollbar, set scrollbar to that position
		if !h.isScrollHorizontal {
			// Horizontal scrollbar not active, let activate first
			h.toggleScrollDirection()
			return
		}
		h.setHorizontalScroll(cx)
		return
	}
}

func (h *view) toggleScrollDirection() {
	log.Infof("toggleScrollDirection")
	if !h.isScrollHorizontal && !h.hasHorizontalScrollbar() {
		// Do not toggle to horizontal if no need for horizontal scroll
		return
	}
	if h.isScrollHorizontal && !h.hasVerticalScrollbar() {
		// Do not toggle to vertical if no need for vertical scroll
		return
	}
	log.Infof("toggle")
	h.isScrollHorizontal = !h.isScrollHorizontal
	h.NotifyChanged()
	if h.properties.OnMoved != nil {
		h.properties.OnMoved()
	}
}

func (h *view) setScrollbarBounds(b Rect) {
	if h.vertScrlView != nil {
		vb := Rect{X: b.W - 3, Y: b.Y, W: b.W - 1, H: b.H}
		if h.guiView.Frame {
			vb.X = vb.X + 1
			vb.W = vb.W + 1
		}
		h.ui.setBounds(h.vertScrlView, vb)
	}
	if h.horzScrlView != nil {
		hb := Rect{X: b.X, Y: b.H - 2, W: b.W, H: b.H - 0}
		h.ui.setBounds(h.horzScrlView, hb)
	}
}

func (h *view) moveVertically(move int) {
	if h.total <= 0 {
		// Cannot scroll empty view
		return
	}
	newCurrent := h.currentIndex + move

	if newCurrent < 0 {
		newCurrent = 0
	}
	if newCurrent >= h.total {
		newCurrent = h.total - 1
	}
	if newCurrent == h.currentIndex {
		// No move, reached top or bottom
		return
	}
	h.currentIndex = newCurrent

	if h.currentIndex < h.firstIndex {
		// Need to scroll view up to the new current line
		h.firstIndex = h.currentIndex
	}
	if h.currentIndex >= h.firstIndex+h.linesCount {
		// Need to scroll view down to the new current line
		h.firstIndex = h.currentIndex - h.linesCount + 1
	}
	h.NotifyChanged()
	if h.properties.OnMoved != nil {
		h.properties.OnMoved()
	}
}

func (h *view) scrollVertically(scroll int) {
	if h.total <= 0 {
		// Cannot scroll empty view
		return
	}
	newFirst := h.firstIndex + scroll

	if newFirst < 0 {
		newFirst = 0
	}
	if newFirst+h.linesCount >= h.total {
		newFirst = h.total - h.linesCount
	}
	if newFirst == h.firstIndex {
		// No move, reached top or bottom
		return
	}
	newCurrent := h.currentIndex + (newFirst - h.firstIndex)

	if newCurrent < newFirst {
		// Need to scroll view up to the new current line
		newCurrent = newFirst
	}
	if newCurrent >= newFirst+h.linesCount {
		// Need to scroll view down to the new current line
		newCurrent = newFirst - h.linesCount - 1
	}

	h.firstIndex = newFirst
	h.currentIndex = newCurrent
	h.NotifyChanged()
	if h.properties.OnMoved != nil {
		h.properties.OnMoved()
	}
}

func (h *view) scrollHorizontal(scroll int) {
	newFirstCharIndex := h.firstCharIndex + scroll
	if newFirstCharIndex < 0 {
		return
	}
	if h.maxLineWidth != 0 && newFirstCharIndex > h.maxLineWidth-h.width/2 {
		return
	}
	h.firstCharIndex = newFirstCharIndex
	h.NotifyChanged()
	if h.properties.OnMoved != nil {
		h.properties.OnMoved()
	}
}

func (h *view) drawVerticalScrollbar(linesCount int) {
	if !h.hasVerticalScrollbar() {
		return
	}
	h.vertScrlView.Clear()
	// Set scrollbar handle color
	color := CMagentaDk
	if h.isScrollHorizontal {
		color = CDark
	}

	sbStart, sbEnd := h.getVerticalScrollbarIndexes()

	// Draw the scrollbar
	var sb strings.Builder
	for i := 0; i < linesCount; i++ {
		if i >= sbStart && i <= sbEnd {
			sb.WriteString(ColorRune(color, scrollBarVerticalHandle))
		} else {
			sb.WriteString(" ")
		}
		sb.WriteString("\n")
	}
	if _, err := h.vertScrlView.Write([]byte(sb.String())); err != nil {
		panic(log.Fatal(err))
	}
}

func (h *view) drawHorizontalScrollbar() {
	if !h.hasHorizontalScrollbar() {
		return
	}
	h.horzScrlView.Clear()
	// Set scrollbar handle color
	color := CMagentaDk
	if !h.isScrollHorizontal {
		color = CDark
	}

	sbStart, sbEnd := h.getHorizontalScrollbarIndexes()

	// Draw the scrollbar
	var sb strings.Builder
	for i := 1; i < h.width-1; i++ {
		if i >= sbStart && i <= sbEnd {
			// Within scrollbar, draw the scrollbar handle
			sb.WriteString(ColorRune(color, scrollBarHorizontalHandle))
		} else {
			sb.WriteString(" ")
		}
	}
	if _, err := h.horzScrlView.Write([]byte(sb.String())); err != nil {
		panic(log.Fatal(err))
	}
}

func (h *view) getVerticalScrollbarIndexes() (start, end int) {
	scrollbarFactor := float64(h.linesCount) / float64(h.total)
	sbStart := int(math.Floor(float64(h.firstIndex) * scrollbarFactor))
	sbSize := int(math.Ceil(float64(h.linesCount) * scrollbarFactor))
	if sbStart+sbSize+1 > h.linesCount {
		sbStart = h.linesCount - sbSize - 1
		if sbStart < 0 {
			sbStart = 0
		}
	}
	if h.linesCount == h.total {
		sbStart = -1
		sbSize = -1
	}
	// log.Infof("sb1: %d, sb2: %d, lines: %d", sbStart, sbSize, h.linesCount)
	return sbStart, sbStart + sbSize
}

func (h *view) getHorizontalScrollbarIndexes() (start, end int) {
	scrollbarFactor := float64(h.width) / float64(h.maxLineWidth)
	sbStart := int(math.Floor(float64(h.firstCharIndex) * scrollbarFactor))
	sbSize := int(math.Ceil(float64(h.width) * scrollbarFactor))
	if sbStart+sbSize+1 > h.width {
		sbStart = h.width - sbSize - 1
		if sbStart < 0 {
			sbStart = 0
		}
	}
	if h.width == h.maxLineWidth {
		sbStart = -1
		sbSize = -1
	}
	// log.Infof("sb1: %d, sb2: %d, chars: %d %d", sbStart, sbSize, h.width, h.maxLineWidth)
	return sbStart, sbStart + sbSize
}

func (h *view) setVerticalScroll(cy int) {
	if !h.hasVerticalScrollbar() {
		return
	}
	setLine := h.total
	if h.height-1 > 0 {
		setLine = int(math.Ceil((float64(cy) / float64(h.height-1)) * float64(h.total)))
	}
	h.scrollVertically(setLine - h.currentIndex)
}

func (h *view) setHorizontalScroll(cx int) {
	if !h.hasHorizontalScrollbar() {
		return
	}
	set := h.maxLineWidth
	if h.width-1 > 0 {
		set = int(math.Ceil((float64(cx) / float64(h.width-1)) * float64(h.maxLineWidth)))
	}
	if set > h.maxLineWidth-h.width/2 {
		set = h.maxLineWidth - h.width/2
	}
	if set < 0 {
		set = 0
	}

	h.firstCharIndex = set
	h.NotifyChanged()
	if h.properties.OnMoved != nil {
		h.properties.OnMoved()
	}
}

func (h *view) hasHorizontalScrollbar() bool {
	return h.horzScrlView != nil && !h.properties.HideHorizontalScrollbar && h.maxLineWidth > h.width
}

func (h *view) hasVerticalScrollbar() bool {
	return h.vertScrlView != nil && !h.properties.HideVerticalScrollbar && h.total > h.height
}
