package cui

import (
	"fmt"
	"strings"

	"github.com/jroimartin/gocui"
	"github.com/michael-reichenauer/gmc/utils/log"
	"golang.org/x/sync/semaphore"
)

const (
	currentLineMarker = '┃' // The marker for current line (left)
)

type BoundFunc func(ww, wh int) Rect

func FullScreen() BoundFunc {
	return func(ww, wh int) Rect { return Rect{0, 0, ww, wh} }
}

func Bounds(b Rect) BoundFunc {
	return func(_, _ int) Rect { return b }
}

func CenterBounds(minWidth, minHeight, maxWidth, maxHeight int) BoundFunc {
	return func(ww, wh int) Rect {
		width := maxWidth
		height := maxHeight
		if maxWidth == 0 {
			width = ww
		}
		if maxHeight == 0 {
			height = wh
		}

		if width > ww {
			width = ww
		}
		if width < minWidth {
			width = minWidth
		}
		if width < 1 {
			width = 1
		}

		if height > wh {
			height = wh
		}
		if height < minHeight {
			height = minHeight
		}
		if height < 1 {
			height = 1
		}

		x := (ww - width) / 2
		y := (wh - height) / 2
		if x < 1 {
			x = 1
		}
		if y < 1 {
			y = 1
		}
		b := Rect{X: x, Y: y, W: width, H: height}
		return b
	}
}

func Relative(bf BoundFunc, relative func(b Rect) Rect) BoundFunc {
	return func(w, h int) Rect {
		vb := bf(w, h)
		return relative(vb)
	}
}

// Properties that adjust view behavior and can be accessed via View.Properties()
type ViewProperties struct {
	Title                   string
	HasFrame                bool
	HideCurrentLineMarker   bool
	HideVerticalScrollbar   bool
	HideHorizontalScrollbar bool

	OnLoad           func()
	OnClose          func()
	OnMouseLeft      func(x, y int)
	OnMouseRight     func(x, y int)
	OnMouseOutside   func()
	OnMoved          func()
	Name             string
	IsEditable       bool
	IsWrap           bool
	IsMoveUpDownWrap bool
	OnEdit           func()
	OnMoveCursor     func()
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
	Show(BoundFunc)
	SetBound(BoundFunc)
	SyncWithView(view View)
	SetCurrentView()
	SetTop()
	SetTitle(title string)
	NotifyChanged()
	SetKey(key interface{}, handler func())
	DeleteKey(key interface{})
	ViewPage() ViewPage
	ReadLines() []string
	SetText(text string)
	Clear()
	PostOnUIThread(func())
	Close()
	ScrollHorizontal(scroll int)
	ScrollVertical(scroll int)
	OnKeyArrowUp()
	OnKeyArrowDown()
	SetCurrentLine(line int)
	ShowLineAtTop(line int, forceScroll bool)
	ShowFrame(isShow bool)
}

type view struct {
	guiView            *gocui.View
	vertScrlView       *gocui.View
	horzScrlView       *gocui.View
	properties         *ViewProperties
	viewData           func(viewPort ViewPage) ViewText
	boundFunc          BoundFunc
	firstIndex         int
	linesCount         int
	currentIndex       int
	total              int
	width              int
	height             int
	ui                 *ui
	isScrollHorizontal bool
	firstCharIndex     int
	notifyThrottler    *semaphore.Weighted
	maxLineWidth       int
	isClosed           bool
	hasWritten         bool
}

func newView(ui *ui, viewData func(viewPort ViewPage) ViewText) *view {
	return &view{
		ui:              ui,
		viewData:        viewData,
		guiView:         ui.createView(),
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
		maxWidth := maxTextWidth(lines)
		firstIndex := viewPort.FirstLine
		if firstIndex > len(lines) {
			firstIndex = len(lines)
		}
		height := viewPort.Height
		if firstIndex+viewPort.Height > len(lines) {
			height = len(lines) - firstIndex
		}
		viewLines := lines[firstIndex : firstIndex+height]

		if viewPort.FirstCharIndex != 0 {
			for i, l := range viewLines {
				if len(l) > viewPort.FirstCharIndex {
					viewLines[i] = l[viewPort.FirstCharIndex:]
				}
			}
		}

		return ViewText{
			Lines:    viewLines,
			Total:    len(lines),
			MaxWidth: maxWidth,
		}
	}
}

func (h *view) SetBound(bf BoundFunc) {
	h.boundFunc = bf
	h.ui.ResizeAllViews()
}

func (h *view) Show(bf BoundFunc) {
	h.boundFunc = bf

	h.ui.addShownView(h)
	log.Debugf("Show %s %s", h.guiView.Name(), h.properties.Name)
	h.guiView.Frame = h.properties.Title != "" || h.properties.HasFrame
	h.guiView.Editable = false
	h.guiView.Wrap = false
	h.guiView.Highlight = false
	if h.properties.Title != "" {
		h.guiView.Title = fmt.Sprintf(" %s ", h.properties.Title)
	}
	if h.properties.IsEditable {
		h.guiView.Editable = true
		h.guiView.Editor = gocui.EditorFunc(h.textEditor)
	}
	if !h.properties.IsEditable {
		h.SetKey(gocui.KeyArrowUp, h.OnKeyArrowUp)
		h.SetKey(gocui.KeyArrowDown, h.OnKeyArrowDown)
		h.SetKey(gocui.KeySpace, h.onKeyPageDown)
	}
	h.guiView.Wrap = h.properties.IsWrap

	h.SetKey(gocui.MouseMiddle, h.toggleScrollDirection)
	h.SetKey(gocui.MouseWheelDown, h.onMouseWheelRollDown)
	h.SetKey(gocui.MouseWheelUp, h.onMouseWheelRollUp)
	h.SetKey(gocui.KeyPgdn, h.onKeyPageDown)
	h.SetKey(gocui.KeyPgup, h.onKeyPageUp)
	h.SetKey(gocui.KeyHome, h.onKeyPageHome)
	h.SetKey(gocui.KeyEnd, h.onKeyPageEnd)

	h.SetKey(gocui.MouseLeft, h.onMouseLeftClick)
	h.SetKey(gocui.MouseRight, h.onMouseRightClick)

	// log.Eventf("ui-view-show", h.Properties().Name)
	if h.properties.OnLoad != nil {
		// Let the actual view handle load to initialise view data
		h.properties.OnLoad()
	}

	if !h.properties.HideVerticalScrollbar {
		h.vertScrlView = h.createVerticalScrollView()
	}
	if !h.properties.HideHorizontalScrollbar {
		h.horzScrlView = h.createHorizontalScrollView()
	}

	mb := h.viewBounds()
	h.setBounds(mb)
	h.NotifyChanged()
}

func (h *view) ShowFrame(isShow bool) {
	h.guiView.Frame = isShow
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
		h.ui.Post(func() {
			h.notifyThrottler.Release(1)
			if h.isClosed {
				return
			}
			if h.properties.IsEditable && h.hasWritten {
				return
			}
			// Clear the view to make room for the new data
			h.guiView.Clear()

			// Get the view size to calculate the view port
			width, height := h.guiView.Size()
			if width <= 1 || height <= 0 {
				// View is to small (not visible)
				return
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
				return
			}

			if h.properties.Title != "" {
				h.guiView.Title = fmt.Sprintf(" %s ", h.properties.Title)
			} else {
				h.guiView.Title = ""
			}

			// Show the new view data for the view port
			if _, err := h.guiView.Write(h.toViewTextBytes(viewData.Lines)); err != nil {
				panic(log.Fatal(err))
			}
			h.hasWritten = true

			h.drawVerticalScrollbar(len(viewData.Lines))
			h.drawHorizontalScrollbar()

			return
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

func (h *view) toViewTextBytes(lines []string) []byte {
	isCurrentView := h == h.ui.currentView()

	var sb strings.Builder
	for i, line := range lines {
		// Draw the current line marker
		if !h.properties.HideCurrentLineMarker && isCurrentView && i+h.firstIndex == h.currentIndex {
			sb.WriteString(ColorRune(CWhite, currentLineMarker))
		} else if !h.properties.HideCurrentLineMarker {
			sb.WriteString(" ")
		}

		// Draw the actual text line
		sb.WriteString(line)
		sb.WriteString("\n")
	}
	return []byte(sb.String())
}

func (h *view) SetTitle(title string) {
	h.guiView.Title = title
}

func (h *view) SetCurrentView() {
	h.ui.setCurrentView(h)
}

func (h *view) SetTop() {
	h.ui.setTop(h.guiView)
	h.ui.setTop(h.vertScrlView)
	h.ui.setTop(h.horzScrlView)
}

func (h *view) SetBottom() {
	h.ui.setBottom(h.guiView)
	h.ui.setBottom(h.vertScrlView)
	h.ui.setBottom(h.horzScrlView)
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
	h.ui.Post(f)
}

func (h *view) Close() {
	log.Debugf("Close %s %s", h.guiView.Name(), h.properties.Name)
	if h.properties.OnClose != nil {
		h.properties.OnClose()
	}
	h.ui.deleteView(h.vertScrlView)
	h.vertScrlView = nil
	h.ui.deleteView(h.horzScrlView)
	h.horzScrlView = nil
	h.ui.closeView(h)
	h.isClosed = true
}

func (h *view) SetKey(key interface{}, handler func()) {
	h.ui.setKey(h.guiView, key, handler)
}

func (h *view) DeleteKey(key interface{}) {
	h.ui.deleteKey(h.guiView, key)
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

func (h *view) SetText(text string) {
	h.Clear()
	if _, err := h.guiView.Write([]byte(text)); err != nil {
		panic(log.Fatal(err))
	}
}

func (h *view) Size() (int, int) {
	return h.guiView.Size()
}

func (h *view) ShowLineAtTop(line int, forceScroll bool) {
	if !forceScroll && line >= h.firstIndex && line <= h.firstIndex+h.linesCount {
		// No need to scroll
		return
	}

	h.ScrollVertical(line - h.firstIndex)
}

func (h *view) OnKeyArrowUp() {
	if h.properties.HideCurrentLineMarker {
		h.ScrollVertical(-1)
		return
	}
	h.moveVertically(-1)
}

func (h *view) OnKeyArrowDown() {
	if h.properties.HideCurrentLineMarker {
		h.ScrollVertical(1)
		return
	}
	h.moveVertically(1)
}

func (h *view) onKeyPageDown() {
	_, y := h.Size()
	h.ScrollVertical(y - 1)
}

func (h *view) onKeyPageUp() {
	_, y := h.Size()
	h.ScrollVertical(-y + 1)
}

func (h *view) onMouseWheelRollDown() {
	if h.isScrollHorizontal {
		h.scrollHorizontal(1)
		return
	}
	h.ScrollVertical(1)
}

func (h *view) onMouseWheelRollUp() {
	if h.isScrollHorizontal {
		h.scrollHorizontal(-1)
		return
	}
	h.ScrollVertical(-1)
}

func (h *view) onKeyPageHome() {
	h.ScrollVertical(-h.currentIndex)
}

func (h *view) onKeyPageEnd() {
	h.ScrollVertical(h.total)
}

func (h *view) onMouseLeftClick() {
	h.mouseDown(h.properties.OnMouseLeft, false)
}

func (h *view) onMouseRightClick() {
	h.mouseDown(h.properties.OnMouseRight, true)
}

func (h *view) mouseDown(mouseHandler func(x, y int), isSetCurrentLine bool) {
	cx, cy := h.guiView.Cursor()

	currentView := h.ui.currentView()
	if h != currentView && currentView.properties.OnMouseOutside != nil {
		// Mouse down, but this is not the current view, inform the current view
		currentView.properties.OnMouseOutside()
		return
	}

	if isSetCurrentLine || mouseHandler == nil {
		// Setting current line to the line that user clicked on
		h.SetCurrentLine(h.firstIndex + cy)
	}

	// Handle mouse down event if mouse custom handler
	if mouseHandler != nil {
		h.PostOnUIThread(func() {
			mouseHandler(cx, cy)
		})
	}
}

func (h *view) SetCurrentLine(line int) {
	h.moveVertically(line - h.currentIndex)
}

func (h *view) viewBounds() Rect {
	ww, wh := h.ui.WindowSize()
	return h.mainBounds(ww, wh)
}

func (h *view) mainBounds(ww, wh int) Rect {
	bounds := h.boundFunc(ww, wh)
	b := Rect{X: bounds.X - 1, Y: bounds.Y - 1, W: bounds.X + bounds.W + 1, H: bounds.Y + bounds.H}
	if b.W < 0 {
		b.W = 0
	}
	if b.H < 0 {
		b.H = 0
	}
	return b
}

func (h *view) resize(width int, height int) {
	b := h.mainBounds(width, height)
	h.setBounds(b)
}

func (h *view) setBounds(b Rect) {
	h.ui.setBounds(h.guiView, b)
	h.setScrollbarBounds(b)
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

func maxTextWidth(lines []string) int {
	maxWidth := 0
	for _, line := range lines {
		if len(line) > maxWidth {
			maxWidth = len(line)
		}
	}
	return maxWidth
}

// instant.
func (h *view) textEditor(v *gocui.View, key gocui.Key, ch rune, mod gocui.Modifier) {
	switch {
	case ch != 0 && mod == 0:
		v.EditWrite(ch)
		h.onEdit()
	case key == gocui.KeySpace:
		v.EditWrite(' ')
		h.onEdit()
	case key == gocui.KeyBackspace || key == gocui.KeyBackspace2:
		v.EditDelete(true)
		h.onEdit()
	case key == gocui.KeyDelete:
		v.EditDelete(false)
		h.onEdit()
	case key == gocui.KeyInsert:
		v.Overwrite = !v.Overwrite
	case key == gocui.KeyEnter:
		v.EditNewLine()
		h.onEdit()
	case key == gocui.KeyArrowDown:
		v.MoveCursor(0, 1, false)
		h.onMoveCursor()
	case key == gocui.KeyArrowUp:
		v.MoveCursor(0, -1, false)
		h.onMoveCursor()
	case key == gocui.KeyArrowLeft:
		v.MoveCursor(-1, 0, false)
		h.onMoveCursor()
	case key == gocui.KeyArrowRight:
		v.MoveCursor(1, 0, false)
		h.onMoveCursor()
	}
}

func (h *view) onEdit() {
	if h.properties.OnEdit != nil {
		h.properties.OnEdit()
	}
}

func (h *view) onMoveCursor() {
	if h.properties.OnMoveCursor != nil {
		h.properties.OnMoveCursor()
	}
}
