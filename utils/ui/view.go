package ui

import (
	"fmt"
	"github.com/jroimartin/gocui"
	"github.com/michael-reichenauer/gmc/utils/log"
)

type Properties struct {
	Title    string
	HasFrame bool

	OnLoad  func()
	OnClose func()
}

type ViewPort struct {
	FirstLine   int
	Lines       int
	CurrentLine int
	Width       int
}

type ViewData struct {
	Text      string
	FirstLine int
	Lines     int
	//CurrentLine int
	TotalLines int
}

type View interface {
	Properties() *Properties
	Show(bounds Rect)
	SetBounds(bounds Rect)
	SetCurrentView()
	NotifyChanged()
	SetKey(key interface{}, modifier gocui.Modifier, handler func())
	CurrentLine() int
	Clear()
	PostOnUIThread(func())
}

type view struct {
	gui     *gocui.Gui
	guiView *gocui.View

	properties *Properties
	viewName   string
	viewData   func(viewPort ViewPort) ViewData
	//currentViewData ViewData
	firstLine   int
	lines       int
	currentLine int
	totalLines  int
}

func newView(ui *UI, viewData func(viewPort ViewPort) ViewData) *view {
	return &view{
		gui:        ui.Gui(),
		viewName:   ui.NewViewName(),
		viewData:   viewData,
		properties: &Properties{}}
}

func (h *view) Show(bounds Rect) {
	if guiView, err := h.gui.SetView(h.viewName, bounds.X-1, bounds.Y-1, bounds.W, bounds.H); err != nil {
		if err != gocui.ErrUnknownView {
			log.Fatal(err)
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

		// Get the view size to calculate the view port
		x, y := h.guiView.Size()
		h.lines = y
		if h.lines <= 0 || x <= 0 {
			// View is to small (not visible)
			return nil
		}
		viewPort := ViewPort{Width: x, FirstLine: h.firstLine, Lines: h.lines, CurrentLine: h.currentLine}

		// Get the view data for that view port and get data sizes (could be smaller than view)
		viewData := h.viewData(viewPort)
		h.firstLine = viewData.FirstLine
		h.lines = viewData.Lines
		h.totalLines = viewData.TotalLines

		// Adjust current line to be in the visible area
		if h.currentLine < h.firstLine {
			h.currentLine = h.firstLine
		}
		if h.currentLine > h.firstLine+h.lines {
			h.currentLine = h.firstLine + h.lines
		}

		if viewData.Lines <= 0 && viewData.Text == "" {
			// No view data
			return nil
		}

		// Show the new view data for the view port
		if _, err := h.guiView.Write([]byte(viewData.Text)); err != nil {
			log.Fatal(err)
		}
		return nil
	})
}

func (h *view) SetBounds(bounds Rect) {
	if _, err := h.gui.SetView(h.viewName, bounds.X-1, bounds.Y-1, bounds.X+bounds.W, bounds.Y+bounds.H); err != nil {
		log.Fatal(err)
	}
}

func (h *view) SetCurrentView() {
	if _, err := h.gui.SetCurrentView(h.viewName); err != nil {
		log.Fatal(err)
	}
}

func (h view) CurrentLine() int {
	return h.currentLine
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
		log.Fatal(err)
	}
}

func (h *view) SetKey(key interface{}, modifier gocui.Modifier, handler func()) {
	if err := h.gui.SetKeybinding(h.viewName, key, modifier, func(gui *gocui.Gui, view *gocui.View) error {
		handler()
		return nil
	}); err != nil {
		log.Fatal(err)
	}
}

func (h *view) Clear() {
	h.guiView.Clear()
}

func (h *view) Cursor() (int, int) {
	return h.guiView.Cursor()
}

func (h *view) SetCursor(x int, y int) error {
	return h.guiView.SetCursor(x, y)
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
	if h.totalLines <= 0 {
		// Cannot scroll empty view
		return
	}
	newCurrent := h.currentLine + move

	if newCurrent < 0 {
		newCurrent = 0
	}
	if newCurrent >= h.totalLines {
		newCurrent = h.totalLines - 1
	}
	if newCurrent == h.currentLine {
		// No move, reached top or bottom
		return
	}

	h.currentLine = newCurrent

	if h.currentLine < h.firstLine {
		// Need to scroll view up to the new current line
		h.firstLine = h.currentLine
	}
	if h.currentLine >= h.firstLine+h.lines {
		// Need to scroll view down to the new current line
		h.firstLine = h.currentLine - h.lines + 1
	}

	h.NotifyChanged()
}

func (h *view) scroll(move int) {
	if h.totalLines <= 0 {
		// Cannot scroll empty view
		return
	}
	newFirst := h.firstLine + move

	if newFirst < 0 {
		newFirst = 0
	}
	if newFirst+h.lines >= h.totalLines {
		newFirst = h.totalLines - h.lines
	}
	if newFirst == h.firstLine {
		// No move, reached top or bottom
		return
	}

	newCurrent := h.currentLine + (newFirst - h.firstLine)

	if newCurrent < newFirst {
		// Need to scroll view up to the new current line
		newCurrent = newFirst
	}
	if newCurrent >= newFirst+h.lines {
		// Need to scroll view down to the new current line
		newCurrent = newFirst - h.lines - 1
	}

	h.firstLine = newFirst
	h.currentLine = newCurrent

	h.NotifyChanged()
}

//func (h *UI) setCursor(gui *gocui.Gui, view *gocui.View, line int) error {
//	log.Infof("Set line %d", line)
//
//	if line >= h.view.viewData.MaxLines {
//		return nil
//	}
//	cx, _ := view.Cursor()
//	_ = view.SetCursor(cx, line)
//
//	h.view.CurrentLine = line
//	if h.view.CurrentLine > h.view.lastLine {
//		move := h.view.CurrentLine - h.view.lastLine
//		h.view.firstLine = h.view.firstLine + move
//		h.view.lastLine = h.view.lastLine + move
//	}
//	h.view.NotifyChanged()
//
//	return nil
//}
