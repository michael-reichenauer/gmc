package ui

import (
	"fmt"
	"github.com/jroimartin/gocui"
	"github.com/michael-reichenauer/gmc/utils/log"
	"golang.org/x/sync/semaphore"
	"math"
	"strings"
)

var (
	currentLineMarker          = '│' // The marker for current line (left)
	scrollBarVerticalHandle    = '▐' // The scrollbar handle (right)
	scrollBarHorizontalHandle  = '▄' // The scrollbar handle (down)
	scrollBarHorizontalHandle2 = '_' // The scrollbar handle inactive (down)
)

type Properties struct {
	Title         string
	HasFrame      bool
	HideCurrent   bool
	HideScrollbar bool

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
}

type ViewPageData struct {
	Lines    []string
	Total    int
	MaxWidth int
}

type Viewer interface {
	Notifier
	Runner
}

type View interface {
	Properties() *Properties
	Show(bounds Rect)
	SetBounds(bounds Rect)
	SetPage(vp ViewPage)
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
	properties         *Properties
	viewName           string
	viewData           func(viewPort ViewPage) ViewPageData
	firstIndex         int
	linesCount         int
	currentIndex       int
	total              int
	width              int
	height             int
	ui                 *UI
	IsScrollHorizontal bool
	FirstCharIndex     int
	notifyThrottler    *semaphore.Weighted
	maxLineWidth       int
}

func newView(ui *UI, viewData func(viewPort ViewPage) ViewPageData) *view {
	return &view{
		ui:              ui,
		viewName:        ui.NewViewName(),
		viewData:        viewData,
		notifyThrottler: semaphore.NewWeighted(int64(1)),
		properties:      &Properties{}}
}

func viewDataFromText(viewText string) func(viewPort ViewPage) ViewPageData {
	return viewDataFromTextFunc(func(viewPort ViewPage) string {
		return viewText
	})
}

func viewDataFromTextFunc(viewText func(viewPort ViewPage) string) func(viewPort ViewPage) ViewPageData {
	return func(viewPort ViewPage) ViewPageData {
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
		return ViewPageData{
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
			h.SetKey(gocui.KeyArrowUp, gocui.ModNone, h.CursorUp)
			h.SetKey(gocui.KeyArrowDown, gocui.ModNone, h.CursorDown)
			h.SetKey(gocui.KeySpace, gocui.ModNone, h.PageDown)
		}

		h.SetKey(gocui.MouseMiddle, gocui.ModNone, h.ToggleScroll)
		h.SetKey(gocui.MouseWheelDown, gocui.ModNone, h.MouseWheelDown)
		h.SetKey(gocui.MouseWheelUp, gocui.ModNone, h.MouseWheelUp)
		h.SetKey(gocui.KeyPgdn, gocui.ModNone, h.PageDown)
		h.SetKey(gocui.KeyPgup, gocui.ModNone, h.PageUp)
		h.SetKey(gocui.KeyHome, gocui.ModNone, h.PageHome)
		h.SetKey(gocui.KeyEnd, gocui.ModNone, h.PageEnd)

		h.SetKey(gocui.MouseLeft, gocui.ModNone, h.MouseLeft)
		h.SetKey(gocui.MouseRight, gocui.ModNone, h.MouseRight)

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

			if !h.properties.HideScrollbar {
				h.drawVerticalScrollbar(len(viewData.Lines))
				h.drawHorizontalScrollbar()
			}
			return nil
		})
	}()
}

func (h *view) SetPage(vp ViewPage) {
	if h.firstIndex != vp.FirstLine ||
		h.currentIndex != vp.CurrentLine ||
		h.FirstCharIndex != vp.FirstCharIndex ||
		h.IsScrollHorizontal != vp.IsHorizontalScrolling {
		h.firstIndex = vp.FirstLine
		h.currentIndex = vp.CurrentLine
		h.FirstCharIndex = vp.FirstCharIndex
		h.IsScrollHorizontal = vp.IsHorizontalScrolling
		h.NotifyChanged()
	}
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
		FirstCharIndex:        h.FirstCharIndex,
		IsHorizontalScrolling: h.IsScrollHorizontal,
	}
}

func (h *view) Properties() *Properties {
	return h.properties
}

func (h *view) PostOnUIThread(f func()) {
	h.ui.gui.Update(func(g *gocui.Gui) error {
		f()
		return nil
	})
}

func (h *view) Close() {
	h.ui.Gui().Cursor = !h.properties.IsEditable
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

func (h *view) CursorUp() {
	h.moveVertically(-1)
}

func (h *view) CursorDown() {
	h.moveVertically(1)
}

func (h *view) MoveLine(line int) {
	h.moveVertically(line)
}

func (h *view) PageDown() {
	_, y := h.Size()
	h.scrollVertically(y - 1)
}

func (h *view) PageUp() {
	_, y := h.Size()
	h.scrollVertically(-y + 1)
}

func (h *view) MouseWheelDown() {
	if h.IsScrollHorizontal {
		h.scrollHorizontal(1)
		return
	}
	h.scrollVertically(1)
}

func (h *view) MouseWheelUp() {
	if h.IsScrollHorizontal {
		h.scrollHorizontal(-1)
		return
	}
	h.scrollVertically(-1)
}

func (h *view) PageHome() {
	h.scrollVertically(-h.currentIndex)
}

func (h *view) PageEnd() {
	h.scrollVertically(h.total)
}

func (h *view) MouseLeft() {
	h.mouseDown(h.properties.OnMouseLeft, false)
}

func (h *view) MouseRight() {
	h.mouseDown(h.properties.OnMouseRight, true)
}

func (h *view) mouseOutside() {
	if h.properties.OnMouseOutside != nil {
		log.Infof("OnMouseOutside handler")
		h.properties.OnMouseOutside()
	}
}

func (h *view) mouseDown(mouseHandler func(x, y int), isMoveLine bool) {
	cx, cy := h.guiView.Cursor()
	if !h.properties.HideScrollbar && cx == h.width-2 {
		h.ScrollVerticalSet(cy)
		return
	}
	if !h.properties.HideScrollbar && cy == h.height-1 {
		h.scrollHorizontalSet(cx)
		return
	}

	if h != h.ui.currentView {
		h.ui.currentView.mouseOutside()
		return
	}

	if isMoveLine || mouseHandler == nil {
		p := h.ViewPage()
		line := p.FirstLine + cy - p.CurrentLine
		h.MoveLine(line)
	}

	if mouseHandler == nil {
		return
	}

	h.PostOnUIThread(func() {
		mouseHandler(cx, cy)
	})
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
	newFirstCharIndex := h.FirstCharIndex + scroll
	if newFirstCharIndex < 0 {
		return
	}
	if h.maxLineWidth != 0 && newFirstCharIndex > h.maxLineWidth-h.width/2 {
		return
	}
	h.FirstCharIndex = newFirstCharIndex
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
	sbStart := int(math.Floor(float64(h.FirstCharIndex) * scrollbarFactor))
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
	log.Infof("sb1: %d, sb2: %d, chars: %d %d", sbStart, sbSize, h.width, h.maxLineWidth)
	return sbStart, sbStart + sbSize
}

func (h *view) ToggleScroll() {
	if !h.IsScrollHorizontal && (h.maxLineWidth == 0 || h.maxLineWidth < h.width) {
		// Do not toggle if no need for horizontal scroll
		return
	}
	h.IsScrollHorizontal = !h.IsScrollHorizontal
	h.NotifyChanged()
	if h.properties.OnMoved != nil {
		h.properties.OnMoved()
	}
}

func (h *view) ScrollVerticalSet(cy int) {
	setLine := h.total
	if h.height-1 > 0 {
		setLine = int(math.Ceil((float64(cy) / float64(h.height-1)) * float64(h.total)))
	}
	h.moveVertically(setLine - h.currentIndex)
}

func (h *view) mainBounds(bounds Rect) Rect {
	b := Rect{X: bounds.X - 1, Y: bounds.Y - 1, W: bounds.X + bounds.W + 1, H: bounds.Y + bounds.H}
	return h.validBounds(b)
}

func (h *view) validBounds(b Rect) Rect {
	if b.W < 0 {
		b.W = 0
	}
	if b.H < 0 {
		b.H = 0
	}
	return b
}

func (h *view) drawVerticalScrollbar(linesCount int) {
	// Remember original values
	x, y := h.guiView.Cursor()
	fg := h.guiView.FgColor

	// Set scrollbar handle color
	h.guiView.FgColor = gocui.ColorMagenta
	if h.IsScrollHorizontal {
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
	if h.maxLineWidth == 0 || h.maxLineWidth < h.width {
		return
	}
	// Remember original values
	x, y := h.guiView.Cursor()
	fg := h.guiView.FgColor

	// Set scrollbar handle color
	h.guiView.FgColor = gocui.ColorMagenta
	handle := scrollBarHorizontalHandle

	if !h.IsScrollHorizontal {
		h.guiView.FgColor = gocui.ColorWhite
		handle = scrollBarHorizontalHandle2
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

func (h *view) scrollHorizontalSet(cx int) {
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
	h.FirstCharIndex = set
	h.NotifyChanged()
	if h.properties.OnMoved != nil {
		h.properties.OnMoved()
	}
}
