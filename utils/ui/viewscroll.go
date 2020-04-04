package ui

import (
	"github.com/jroimartin/gocui"
	"math"
)

const (
	scrollBarVerticalHandle   = '┃' // The vertical scrollbar handle (right)
	scrollBarHorizontalHandle = '━' // The horizontal scrollbar handle (down)
)

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

func (h *view) drawVerticalScrollbar(linesCount int) {
	if h.properties.HideVerticalScrollbar || h.total < h.height {
		return
	}
	// Remember original values
	x, y := h.guiView.Cursor()
	fg := h.guiView.FgColor

	// Set scrollbar handle color
	h.guiView.FgColor = gocui.ColorMagenta
	if h.isScrollHorizontal {
		h.guiView.FgColor = gocui.ColorWhite
	}

	sx := h.width - 1
	sbStart, sbEnd := h.getVerticalScrollbarIndexes()

	// Draw the scrollbar
	for i := 0; i < linesCount; i++ {
		_ = h.guiView.SetCursor(sx, i)
		h.guiView.EditDelete(true)
		if i >= sbStart && i <= sbEnd {
			// Within scrollbar, draw the scrollbar handle
			h.guiView.EditWrite(scrollBarVerticalHandle)
		} else {
			h.guiView.EditWrite(' ')
		}
	}

	// Restore values
	_ = h.guiView.SetCursor(x, y)
	h.guiView.FgColor = fg
}

func (h *view) drawHorizontalScrollbar() {
	if h.properties.HideHorizontalScrollbar || h.maxLineWidth == 0 || h.maxLineWidth < h.width {
		return
	}
	// Remember original values
	x, y := h.guiView.Cursor()
	fg := h.guiView.FgColor

	// Set scrollbar handle color
	h.guiView.FgColor = gocui.ColorMagenta
	handle := scrollBarHorizontalHandle

	if !h.isScrollHorizontal {
		h.guiView.FgColor = gocui.ColorWhite
		handle = scrollBarHorizontalHandle
	}

	sy := h.height - 1
	sbStart, sbEnd := h.getHorizontalScrollbarIndexes()

	// Draw the scrollbar
	for i := 1; i < h.width-1; i++ {
		_ = h.guiView.SetCursor(i, sy)
		h.guiView.EditDelete(true)
		if i >= sbStart && i <= sbEnd {
			// Within scrollbar, draw the scrollbar handle
			h.guiView.EditWrite(handle)
		} else {
			h.guiView.EditWrite(' ')
		}
	}

	// Restore values
	_ = h.guiView.SetCursor(x, y)
	h.guiView.FgColor = fg
}

func (h *view) setVerticalScroll(cy int) {
	setLine := h.total
	if h.height-1 > 0 {
		setLine = int(math.Ceil((float64(cy) / float64(h.height-1)) * float64(h.total)))
	}
	h.scrollVertically(setLine - h.currentIndex)
}

func (h *view) setHorizontalScroll(cx int) {
	if h.maxLineWidth == 0 {
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
