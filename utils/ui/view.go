package ui

import (
	"fmt"
	"github.com/jroimartin/gocui"
	"github.com/michael-reichenauer/gmc/utils/log"
	"math"
	"strings"
)

var (
	currentLineMarker = '│'
	scrollBarHandle   = "▐"
	scrollBarHandle2  = "█"
)

type Properties struct {
	Title       string
	HasFrame    bool
	HideCurrent bool

	OnLoad  func()
	OnClose func()
	Name    string
}

type ViewPage struct {
	Width       int
	Height      int
	FirstLine   int
	CurrentLine int
}

type ViewData struct {
	Lines      []string
	FirstIndex int
	Total      int
}

type View interface {
	Properties() *Properties
	Show(bounds Rect)
	SetBounds(bounds Rect)
	SetCurrentView()
	SetTop()
	SetBottom()
	SetTitle(title string)
	NotifyChanged()
	SetKey(key interface{}, modifier gocui.Modifier, handler func())
	ViewPage() ViewPage
	Clear()
	PostOnUIThread(func())
	Close()
}

type view struct {
	gui     *gocui.Gui
	guiView *gocui.View

	properties   *Properties
	viewName     string
	viewData     func(viewPort ViewPage) ViewData
	firstIndex   int
	linesCount   int
	currentIndex int
	total        int
	width        int
}

func newView(ui *UI, viewData func(viewPort ViewPage) ViewData) *view {
	return &view{
		gui:        ui.Gui(),
		viewName:   ui.NewViewName(),
		viewData:   viewData,
		properties: &Properties{}}
}

func (h *view) Show(bounds Rect) {
	if guiView, err := h.gui.SetView(h.viewName, bounds.X-1, bounds.Y-1, bounds.W, bounds.H); err != nil {
		if err != gocui.ErrUnknownView {
			panic(log.Fatalf(err, "%s %d,%d,%d,%d", h.viewName, bounds.X-1, bounds.Y-1, bounds.W, bounds.H))
		}

		h.guiView = guiView
		h.guiView.Frame = h.properties.Title != "" || h.properties.HasFrame
		h.guiView.Editable = false
		h.guiView.Wrap = false
		h.guiView.Highlight = false
		h.guiView.SelBgColor = gocui.ColorBlue
		if h.properties.Title != "" {
			h.guiView.Title = fmt.Sprintf(" %s ", h.properties.Title)
		}

		h.SetKey(gocui.KeyArrowDown, gocui.ModNone, h.CursorDown)
		h.SetKey(gocui.KeySpace, gocui.ModNone, h.PageDown)
		h.SetKey(gocui.KeyPgdn, gocui.ModNone, h.PageDown)
		h.SetKey(gocui.KeyPgup, gocui.ModNone, h.PageUpp)
		h.SetKey(gocui.KeyArrowUp, gocui.ModNone, h.CursorUp)

		log.Eventf("ui-view-show", h.Properties().Name)
		if h.properties.OnLoad != nil {
			// Let the actual view handle load to initialise view data
			h.properties.OnLoad()
		}
	}
}

func (h *view) NotifyChanged() {
	h.gui.Update(func(g *gocui.Gui) error {
		// Clear the view to make room for the new data
		h.guiView.Clear()
		isCurrent := h.gui.CurrentView() == h.guiView

		// Get the view size to calculate the view port
		width, height := h.guiView.Size()
		if width <= 1 || height <= 0 {
			// View is to small (not visible)
			return nil
		}
		viewPort := h.ViewPage()

		// Get the view data for that view port and get data sizes (could be smaller than view)
		viewData := h.viewData(viewPort)

		h.width = width
		h.firstIndex = viewData.FirstIndex
		h.total = viewData.Total
		h.linesCount = len(viewData.Lines)
		if h.total < len(viewData.Lines) {
			// total was probably not specified (or wrong), lets adjust
			h.total = len(viewData.Lines)
		}
		if h.linesCount > height {
			// view data lines are more than view height, lets skip some lines
			h.linesCount = height
			viewData.Lines = viewData.Lines[:height]
		}

		// Adjust current line to be in the visible area
		if h.currentIndex < h.firstIndex {
			h.currentIndex = h.firstIndex
		}
		if h.currentIndex > h.firstIndex+h.linesCount {
			h.currentIndex = h.firstIndex + h.linesCount
		}

		if h.linesCount == 0 {
			// No view data
			return nil
		}

		// Show the new view data for the view port
		if _, err := h.guiView.Write(h.toViewTextBytes(viewData.Lines, isCurrent)); err != nil {
			panic(log.Fatal(err))
		}
		return nil
	})
}

func (h *view) toViewTextBytes(lines []string, idCurrent bool) []byte {
	sbStart, sbEnd := h.getScrollbarIndexes()

	var sb strings.Builder
	for i, line := range lines {
		// Draw the current line marker
		if !h.properties.HideCurrent && idCurrent && i+h.firstIndex == h.currentIndex {
			sb.WriteString(ColorRune(CWhite, currentLineMarker))
		} else {
			sb.WriteString(" ")
		}

		// Draw the actual text line
		sb.WriteString(line)

		// // Draw the scrollbar
		if i >= sbStart && i <= sbEnd {
			// Within scrollbar, draw the scrollbar handle
			sb.WriteString(MagentaDk(scrollBarHandle))
		} else {
			sb.WriteString(" ")
		}

		sb.WriteString("\n")
	}
	return []byte(sb.String())
}

func (h *view) SetBounds(bounds Rect) {
	if _, err := h.gui.SetView(h.viewName, bounds.X-1, bounds.Y-1, bounds.X+bounds.W, bounds.Y+bounds.H); err != nil {
		panic(log.Fatal(err))
	}
}

func (h *view) SetTitle(title string) {
	h.guiView.Title = title
}

func (h *view) SetCurrentView() {
	if _, err := h.gui.SetCurrentView(h.viewName); err != nil {
		panic(log.Fatal(err))
	}
}

func (h *view) SetTop() {
	if _, err := h.gui.SetViewOnTop(h.viewName); err != nil {
		panic(log.Fatal(err))
	}
}

func (h *view) SetBottom() {
	if _, err := h.gui.SetViewOnBottom(h.viewName); err != nil {
		panic(log.Fatal(err))
	}
}

func (h view) ViewPage() ViewPage {
	// Get the view size to calculate the view port
	width, height := h.guiView.Size()
	if width <= 1 || height <= 0 {
		// View is to small (not visible)
		return ViewPage{Width: 1, FirstLine: 0, Height: 1, CurrentLine: 0}
	}
	return ViewPage{Width: width - 2, FirstLine: h.firstIndex, Height: height, CurrentLine: h.currentIndex}
}

func (h *view) Properties() *Properties {
	return h.properties
}

func (h *view) PostOnUIThread(f func()) {
	h.gui.Update(func(g *gocui.Gui) error {
		f()
		return nil
	})
}

func (h *view) Close() {
	if h.properties.OnClose != nil {
		h.properties.OnClose()
	}
	if err := h.gui.DeleteView(h.viewName); err != nil {
		panic(log.Fatal(err))
	}
}

func (h *view) SetKey(key interface{}, modifier gocui.Modifier, handler func()) {
	if err := h.gui.SetKeybinding(h.viewName, key, modifier, func(gui *gocui.Gui, view *gocui.View) error {
		handler()
		return nil
	}); err != nil {
		panic(log.Fatal(err))
	}
}

func (h *view) Clear() {
	h.guiView.Clear()
}

func (h *view) Size() (int, int) {
	return h.guiView.Size()
}

func (h *view) CursorUp() {
	h.move(-1)
}

func (h *view) CursorDown() {
	h.move(1)
}

func (h *view) PageDown() {
	_, y := h.Size()
	h.scroll(y - 1)
}

func (h *view) PageUpp() {
	_, y := h.Size()
	h.scroll(-y + 1)
}

func (h *view) move(move int) {
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
}

func (h *view) scroll(move int) {
	if h.total <= 0 {
		// Cannot scroll empty view
		return
	}
	newFirst := h.firstIndex + move

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
}

func (h *view) getScrollbarIndexes() (start, end int) {
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
