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
	Lines      []string
	FirstIndex int
	Total      int
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
	ViewPage() ViewPage
	Clear()
	PostOnUIThread(func())
	Close()
	ScrollHorizontal(scroll int)
}

type view struct {
	guiView    *gocui.View
	scrollView *gocui.View

	properties         *Properties
	viewName           string
	viewData           func(viewPort ViewPage) ViewPageData
	firstIndex         int
	linesCount         int
	currentIndex       int
	total              int
	width              int
	ui                 *UI
	IsScrollHorizontal bool
	FirstCharIndex     int
}

func newViewFromPageFunc(ui *UI, viewData func(viewPort ViewPage) ViewPageData) *view {
	return &view{
		ui:         ui,
		viewName:   ui.NewViewName(),
		viewData:   viewData,
		properties: &Properties{}}
}

func newViewFromTextFunc(ui *UI, viewText func(viewPort ViewPage) string) *view {
	return &view{
		ui:         ui,
		viewName:   ui.NewViewName(),
		viewData:   viewDataFromTextFunc(viewText),
		properties: &Properties{}}
}

func newView(ui *UI, text string) *view {
	return &view{
		ui:         ui,
		viewName:   ui.NewViewName(),
		viewData:   viewDataFromText(text),
		properties: &Properties{}}
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
			Lines:      lines,
			FirstIndex: viewPort.FirstLine,
			Total:      len(lines),
		}
	}
}

func (h *view) Show(bounds Rect) {
	if guiView, err := h.ui.gui.SetView(h.viewName, bounds.X-1, bounds.Y-1, bounds.X+bounds.W, bounds.Y+bounds.H); err != nil {
		if err != gocui.ErrUnknownView {
			panic(log.Fatalf(err, "%s %d,%d,%d,%d", h.viewName, bounds.X-1, bounds.Y-1, bounds.W, bounds.H))
		}

		h.guiView = guiView
		h.guiView.Frame = h.properties.Title != "" || h.properties.HasFrame
		h.guiView.Editable = false
		h.guiView.Wrap = false
		h.guiView.Highlight = false
		if h.properties.Title != "" {
			h.guiView.Title = fmt.Sprintf(" %s ", h.properties.Title)
		}

		h.SetKey(gocui.KeyArrowUp, gocui.ModNone, h.CursorUp)
		h.SetKey(gocui.KeyArrowDown, gocui.ModNone, h.CursorDown)
		h.SetKey(gocui.MouseMiddle, gocui.ModNone, h.ToggleScroll)
		h.SetKey(gocui.MouseWheelDown, gocui.ModNone, h.MouseWheelDown)
		h.SetKey(gocui.MouseWheelUp, gocui.ModNone, h.MouseWheelUp)
		h.SetKey(gocui.KeySpace, gocui.ModNone, h.PageDown)
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
	if scrollView, err := h.ui.gui.SetView(h.scrlName(), bounds.X+bounds.W-2, bounds.Y-1, bounds.X+bounds.W, bounds.Y+bounds.H); err != nil {
		if err != gocui.ErrUnknownView {
			panic(log.Fatalf(err, "%s", h.scrlName()))
		}
		h.scrollView = scrollView
		h.scrollView.Frame = false
		h.scrollView.Editable = false
		h.scrollView.Wrap = false
		h.scrollView.Highlight = false
		h.scrollView.Title = ""
	}
}

func (h *view) ScrollHorizontal(scroll int) {
	h.scrollHorizontal(scroll)
}

func (h *view) NotifyChanged() {
	h.ui.gui.Update(func(g *gocui.Gui) error {
		// Clear the view to make room for the new data
		h.guiView.Clear()
		h.scrollView.Clear()

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
			if _, err := h.scrollView.Write(h.toScrollTextBytes(len(viewData.Lines))); err != nil {
				panic(log.Fatal(err))
			}
		}
		return nil
	})
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

func (h *view) toScrollTextBytes(linesCount int) []byte {
	sbStart, sbEnd := h.getScrollbarIndexes()

	var sb strings.Builder
	for i := 0; i < linesCount; i++ {
		// // Draw the scrollbar
		if i >= sbStart && i <= sbEnd {
			// Within scrollbar, draw the scrollbar handle
			if h.IsScrollHorizontal {
				sb.WriteString(Dark(scrollBarHandle))
			} else {
				sb.WriteString(MagentaDk(scrollBarHandle))
			}

		} else {
			sb.WriteString(" ")
		}
		sb.WriteString("\n")
	}
	return []byte(sb.String())
}

func (h *view) SetBounds(bounds Rect) {
	if _, err := h.ui.gui.SetView(h.viewName, bounds.X-1, bounds.Y-1, bounds.X+bounds.W, bounds.Y+bounds.H); err != nil {
		panic(log.Fatal(err))
	}
	if _, err := h.ui.gui.SetView(h.scrlName(), bounds.X+bounds.W-2, bounds.Y-1, bounds.X+bounds.W, bounds.Y+bounds.H); err != nil {
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
	if _, err := h.ui.gui.SetViewOnTop(h.scrlName()); err != nil {
		panic(log.Fatal(err))
	}
}

func (h *view) SetBottom() {
	if _, err := h.ui.gui.SetViewOnBottom(h.viewName); err != nil {
		panic(log.Fatal(err))
	}
	if _, err := h.ui.gui.SetViewOnBottom(h.scrlName()); err != nil {
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
		Width:                 width - 2,
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
	if h.properties.OnClose != nil {
		h.properties.OnClose()
	}
	if err := h.ui.gui.DeleteView(h.scrlName()); err != nil {
		panic(log.Fatal(err))
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

func (h *view) MoveLine(line int) {
	h.move(line)
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
	log.Infof("Cursor %d,%d for %q", cx, cy, h.viewName)

	if h != h.ui.currentView {
		log.Infof("Mouse outside for %s %q", h.viewName, h.ui.currentView.viewName)
		h.ui.currentView.mouseOutside()
		return
	}

	if isMoveLine || mouseHandler == nil {
		p := h.ViewPage()
		line := cy - p.CurrentLine
		log.Infof("Mouse move %d lines", line)
		h.MoveLine(line)
	}

	if mouseHandler == nil {
		return
	}

	h.PostOnUIThread(func() {
		log.Infof("Mouse handler %d, %d", cx, cy)
		mouseHandler(cx, cy)
	})
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
	h.FirstCharIndex = newFirstCharIndex
	h.NotifyChanged()
	if h.properties.OnMoved != nil {
		h.properties.OnMoved()
	}
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

func (h *view) ToggleScroll() {
	h.IsScrollHorizontal = !h.IsScrollHorizontal
	h.NotifyChanged()
	if h.properties.OnMoved != nil {
		h.properties.OnMoved()
	}
}

func (h *view) scrlName() string {
	return h.viewName + "scrl"
}
