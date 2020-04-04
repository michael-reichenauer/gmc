package ui

import (
	"fmt"
	"github.com/jroimartin/gocui"
	"github.com/michael-reichenauer/gmc/utils/log"
	"golang.org/x/sync/semaphore"
	"strings"
)

const (
	currentLineMarker = '│' // The marker for current line (left)
)

// Properties that adjust view behavior and can be accessed via View.Properties()
type ViewProperties struct {
	Title                   string
	HasFrame                bool
	HideCurrent             bool
	HideVerticalScrollbar   bool
	HideHorizontalScrollbar bool

	OnLoad         func()
	OnClose        func()
	OnMouseLeft    func(x, y int)
	OnMouseRight   func(x, y int)
	OnMouseOutside func()
	OnMoved        func()
	Name           string
	IsEditable     bool
}

type ViewPage struct {
	Width                 int
	Height                int
	FirstLine             int
	CurrentLine           int
	FirstCharIndex        int
	IsHorizontalScrolling bool
	MaxLineWidth          int
}

type ViewText struct {
	Lines    []string
	Total    int
	MaxWidth int
}

type Viewer interface {
	Notifier
	Runner
}

type View interface {
	Properties() *ViewProperties
	Show(bounds Rect)
	SetBounds(bounds Rect)
	SyncWithView(view View)
	SetCurrentView()
	SetTop()
	SetBottom()
	SetTitle(title string)
	NotifyChanged()
	SetKey(key interface{}, modifier gocui.Modifier, handler func())
	DeleteKey(key interface{}, modifier gocui.Modifier)
	ViewPage() ViewPage
	Clear()
	PostOnUIThread(func())
	Close()
	ScrollHorizontal(scroll int)
}

type view struct {
	guiView            *gocui.View
	properties         *ViewProperties
	viewName           string
	viewData           func(viewPort ViewPage) ViewText
	firstIndex         int
	linesCount         int
	currentIndex       int
	total              int
	width              int
	height             int
	ui                 *UI
	isScrollHorizontal bool
	firstCharIndex     int
	notifyThrottler    *semaphore.Weighted
	maxLineWidth       int
}

func newView(ui *UI, viewData func(viewPort ViewPage) ViewText) *view {
	return &view{
		ui:              ui,
		viewName:        ui.NewViewName(),
		viewData:        viewData,
		notifyThrottler: semaphore.NewWeighted(int64(1)),
		properties:      &ViewProperties{}}
}

func viewDataFromText(viewText string) func(viewPort ViewPage) ViewText {
	return viewDataFromTextFunc(func(viewPort ViewPage) string {
		return viewText
	})
}

func viewDataFromTextFunc(viewText func(viewPort ViewPage) string) func(viewPort ViewPage) ViewText {
	return func(viewPort ViewPage) ViewText {
		lines := strings.Split(viewText(viewPort), "\n")
		firstIndex := viewPort.FirstLine
		if firstIndex > len(lines) {
			firstIndex = len(lines)
		}
		height := viewPort.Height
		if firstIndex+viewPort.Height > len(lines) {
			height = len(lines) - firstIndex
		}
		lines = lines[firstIndex : firstIndex+height]
		return ViewText{
			Lines: lines,
			Total: len(lines),
		}
	}
}

func (h *view) Show(bounds Rect) {
	h.ui.Gui().Cursor = h.properties.IsEditable
	mb := h.mainBounds(bounds)

	if guiView, err := h.ui.gui.SetView(h.viewName, mb.X, mb.Y, mb.W, mb.H); err != nil {
		if err != gocui.ErrUnknownView {
			panic(log.Fatalf(err, "%s %+v,%d,%d,%d", h.viewName, mb))
		}

		h.guiView = guiView
		h.guiView.Frame = h.properties.Title != "" || h.properties.HasFrame
		h.guiView.Editable = false
		h.guiView.Wrap = false
		h.guiView.Highlight = false
		if h.properties.Title != "" {
			h.guiView.Title = fmt.Sprintf(" %s ", h.properties.Title)
		}
		if h.properties.IsEditable {
			h.guiView.Editable = true
		}
		if !h.properties.IsEditable {
			h.SetKey(gocui.KeyArrowUp, gocui.ModNone, h.onKeyArrowUp)
			h.SetKey(gocui.KeyArrowDown, gocui.ModNone, h.onKeyArrowDown)
			h.SetKey(gocui.KeySpace, gocui.ModNone, h.onKeyPageDown)
		}

		h.SetKey(gocui.MouseMiddle, gocui.ModNone, h.toggleScrollDirection)
		h.SetKey(gocui.MouseWheelDown, gocui.ModNone, h.onMouseWheelRollDown)
		h.SetKey(gocui.MouseWheelUp, gocui.ModNone, h.onMouseWheelRollUp)
		h.SetKey(gocui.KeyPgdn, gocui.ModNone, h.onKeyPageDown)
		h.SetKey(gocui.KeyPgup, gocui.ModNone, h.onKeyPageUp)
		h.SetKey(gocui.KeyHome, gocui.ModNone, h.onKeyPageHome)
		h.SetKey(gocui.KeyEnd, gocui.ModNone, h.onKeyPageEnd)

		h.SetKey(gocui.MouseLeft, gocui.ModNone, h.onMouseLeftClick)
		h.SetKey(gocui.MouseRight, gocui.ModNone, h.onMouseRightClick)

		log.Eventf("ui-view-show", h.Properties().Name)
		if h.properties.OnLoad != nil {
			// Let the actual view handle load to initialise view data
			h.properties.OnLoad()
		}
	}
}

func (h *view) ScrollHorizontal(scroll int) {
	h.scrollHorizontal(scroll)
}

func (h *view) NotifyChanged() {
	if !h.notifyThrottler.TryAcquire(1) {
		// Already scheduled notify, skipping this
		return
	}
	go func() {
		h.ui.gui.Update(func(g *gocui.Gui) error {
			h.notifyThrottler.Release(1)

			// Clear the view to make room for the new data
			h.guiView.Clear()

			isCurrent := h.ui.gui.CurrentView() == h.guiView

			// Get the view size to calculate the view port
			width, height := h.guiView.Size()
			if width <= 1 || height <= 0 {
				// View is to small (not visible)
				return nil
			}
			viewPort := h.ViewPage()

			// Get the view data for that view port and get data sizes (could be smaller than view)
			viewData := h.viewData(viewPort)
			if viewData.Total < len(viewData.Lines) {
				viewData.Total = len(viewData.Lines)
			}
			h.maxLineWidth = viewData.MaxWidth

			h.width = width
			h.height = height
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

			if h.properties.Title != "" {
				h.guiView.Title = fmt.Sprintf(" %s ", h.properties.Title)
			} else {
				h.guiView.Title = ""
			}

			// Show the new view data for the view port
			if _, err := h.guiView.Write(h.toViewTextBytes(viewData.Lines, isCurrent)); err != nil {
				panic(log.Fatal(err))
			}

			if !h.properties.HideVerticalScrollbar {
				h.drawVerticalScrollbar(len(viewData.Lines))
			}
			if !h.properties.HideHorizontalScrollbar {
				h.drawHorizontalScrollbar()
			}
			return nil
		})
	}()
}

func (h *view) SyncWithView(v View) {
	p := v.ViewPage()
	h.firstIndex = p.FirstLine
	h.currentIndex = p.CurrentLine
	h.firstCharIndex = p.FirstCharIndex
	h.isScrollHorizontal = p.IsHorizontalScrolling
	h.maxLineWidth = p.MaxLineWidth
	h.NotifyChanged()
}

func (h *view) toViewTextBytes(lines []string, idCurrent bool) []byte {
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
		sb.WriteString("\n")
	}
	return []byte(sb.String())
}

func (h *view) SetBounds(bounds Rect) {
	mb := h.mainBounds(bounds)
	if _, err := h.ui.gui.SetView(h.viewName, mb.X, mb.Y, mb.W, mb.H); err != nil {
		panic(log.Fatal(err))
	}
}

func (h *view) SetTitle(title string) {
	h.guiView.Title = title
}

func (h *view) SetCurrentView() {
	h.ui.SetCurrentView(h)
}

func (h *view) SetTop() {
	if _, err := h.ui.gui.SetViewOnTop(h.viewName); err != nil {
		panic(log.Fatal(err))
	}
}

func (h *view) SetBottom() {
	if _, err := h.ui.gui.SetViewOnBottom(h.viewName); err != nil {
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
	return ViewPage{
		Width:                 width - 3,
		FirstLine:             h.firstIndex,
		Height:                height,
		CurrentLine:           h.currentIndex,
		FirstCharIndex:        h.firstCharIndex,
		IsHorizontalScrolling: h.isScrollHorizontal,
		MaxLineWidth:          h.maxLineWidth,
	}
}

func (h *view) Properties() *ViewProperties {
	return h.properties
}

func (h *view) PostOnUIThread(f func()) {
	h.ui.gui.Update(func(g *gocui.Gui) error {
		f()
		return nil
	})
}

func (h *view) Close() {
	h.ui.Gui().Cursor = false
	if h.properties.OnClose != nil {
		h.properties.OnClose()
	}
	if err := h.ui.gui.DeleteView(h.viewName); err != nil {
		panic(log.Fatal(err))
	}
}

func (h *view) SetKey(key interface{}, modifier gocui.Modifier, handler func()) {
	if err := h.ui.gui.SetKeybinding(h.viewName, key, modifier, func(gui *gocui.Gui, view *gocui.View) error {
		handler()
		return nil
	}); err != nil {
		panic(log.Fatal(err))
	}
}

func (h *view) DeleteKey(key interface{}, modifier gocui.Modifier) {
	if err := h.ui.gui.DeleteKeybinding(h.viewName, key, modifier); err != nil {
		panic(log.Fatal(err))
	}
}

func (h *view) Clear() {
	h.guiView.Clear()
}

func (h *view) Read() string {
	return strings.Join(h.ReadLines(), "\n")
}

func (h *view) ReadLines() []string {
	return h.guiView.BufferLines()
}

func (h *view) Size() (int, int) {
	return h.guiView.Size()
}

func (h *view) onKeyArrowUp() {
	h.moveVertically(-1)
}

func (h *view) onKeyArrowDown() {
	h.moveVertically(1)
}

func (h *view) onKeyPageDown() {
	_, y := h.Size()
	h.scrollVertically(y - 1)
}

func (h *view) onKeyPageUp() {
	_, y := h.Size()
	h.scrollVertically(-y + 1)
}

func (h *view) onMouseWheelRollDown() {
	if h.isScrollHorizontal {
		h.scrollHorizontal(1)
		return
	}
	h.scrollVertically(1)
}

func (h *view) onMouseWheelRollUp() {
	if h.isScrollHorizontal {
		h.scrollHorizontal(-1)
		return
	}
	h.scrollVertically(-1)
}

func (h *view) onKeyPageHome() {
	h.scrollVertically(-h.currentIndex)
}

func (h *view) onKeyPageEnd() {
	h.scrollVertically(h.total)
}

func (h *view) onMouseLeftClick() {
	h.mouseDown(h.properties.OnMouseLeft, false)
}

func (h *view) onMouseRightClick() {
	h.mouseDown(h.properties.OnMouseRight, true)
}

func (h *view) mouseDown(mouseHandler func(x, y int), isSetCurrentLine bool) {
	cx, cy := h.guiView.Cursor()

	if h != h.ui.currentView && h.ui.currentView.properties.OnMouseOutside != nil {
		// Mouse down, but this is not the current view, inform the current view
		h.ui.currentView.properties.OnMouseOutside()
		return
	}

	if !h.properties.HideVerticalScrollbar && cx == h.width-2 {
		// Mouse down in vertical scrollbar, set scrollbar to that position
		h.setVerticalScroll(cy)
		return
	}
	if !h.properties.HideHorizontalScrollbar && cy == h.height-1 {
		// Mouse down in horizontal scrollbar, set scrollbar to that position
		h.setHorizontalScroll(cx)
		return
	}

	if isSetCurrentLine || mouseHandler == nil {
		// Setting current line to the line that user clicked on
		p := h.ViewPage()
		line := p.FirstLine + cy - p.CurrentLine
		h.scrollVertically(line)
	}

	// Handle mouse down event if mouse custom handler
	if mouseHandler != nil {
		h.PostOnUIThread(func() {
			mouseHandler(cx, cy)
		})
	}
}

func (h *view) toggleScrollDirection() {
	if !h.isScrollHorizontal && (h.maxLineWidth == 0 || h.maxLineWidth < h.width) {
		// Do not toggle to horizontal if no need for horizontal scroll
		return
	}

	h.isScrollHorizontal = !h.isScrollHorizontal
	h.NotifyChanged()
	if h.properties.OnMoved != nil {
		h.properties.OnMoved()
	}
}

func (h *view) mainBounds(bounds Rect) Rect {
	b := Rect{X: bounds.X - 1, Y: bounds.Y - 1, W: bounds.X + bounds.W + 1, H: bounds.Y + bounds.H}
	if b.W < 0 {
		b.W = 0
	}
	if b.H < 0 {
		b.H = 0
	}
	return b
}
