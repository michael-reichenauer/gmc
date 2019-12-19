package ui

import (
	"fmt"
	"github.com/jroimartin/gocui"
	"github.com/michael-reichenauer/gmc/utils/log"
)

type Properties struct {
	Title    string
	HasFrame bool

	OnViewData func(viewPort ViewPort) ViewData
	OnLoad     func()
	OnClose    func()
}

type ViewData struct {
	Text     string
	MaxLines int
	First    int
	Last     int
	Current  int
}

type ViewPort struct {
	Width   int
	First   int
	Last    int
	Current int
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

	properties  *Properties
	viewName    string
	viewData    ViewData
	firstLine   int
	lastLine    int
	currentLine int
}

func newView(ui *UI) *view {
	return &view{
		gui:        ui.Gui(),
		viewName:   ui.NewViewName(),
		properties: &Properties{}}
}

func (h *view) Show(bounds Rect) {
	log.Infof("show before set")
	if gv, err := h.gui.SetView(h.viewName, bounds.X-1, bounds.Y-1, bounds.W+1, bounds.H+1); err != nil {
		if err != gocui.ErrUnknownView {
			log.Fatal(err)
		}

		h.guiView = gv
		_, vy := h.guiView.Size()
		h.firstLine = 0
		h.lastLine = vy - 1

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
			h.properties.OnLoad()
		}
	}
	log.Infof("after more2")
}

func (h *view) SetBounds(bounds Rect) {
	if _, err := h.gui.SetView(h.viewName, bounds.X-1, bounds.Y-1, bounds.W+1, bounds.H+1); err != nil {
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

func (h *view) NotifyChanged() {
	h.gui.Update(func(g *gocui.Gui) error {
		h.guiView.Clear()
		x, y := h.guiView.Size()
		h.lastLine = h.firstLine + y - 1
		h.viewData = h.properties.OnViewData(ViewPort{Width: x, First: h.firstLine, Last: h.lastLine, Current: h.currentLine})
		h.firstLine = h.viewData.First
		h.lastLine = h.viewData.Last
		h.currentLine = h.viewData.Current
		if _, err := h.guiView.Write([]byte(h.viewData.Text)); err != nil {
			log.Fatal(err)
		}
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
	if h.currentLine <= 0 {
		return
	}
	h.currentLine = h.currentLine - 1
	if h.currentLine < h.firstLine {
		move := h.firstLine - h.currentLine
		h.firstLine = h.firstLine - move
		h.lastLine = h.lastLine - move
	}
	h.NotifyChanged()
}

func (h *view) CursorDown() {
	if h.currentLine >= h.viewData.MaxLines-1 {
		return
	}
	h.currentLine = h.currentLine + 1
	if h.currentLine > h.lastLine {
		move := h.currentLine - h.lastLine
		h.firstLine = h.firstLine + move
		h.lastLine = h.lastLine + move
	}
	h.NotifyChanged()
}
func (h *view) PageDown() {
	_, y := h.Size()
	move := y - 2
	if h.lastLine+move >= h.viewData.MaxLines-1 {
		move = h.viewData.MaxLines - 1 - h.lastLine
	}
	if move < 1 {
		return
	}
	h.firstLine = h.firstLine + move
	h.lastLine = h.lastLine + move
	h.currentLine = h.currentLine + move
	h.NotifyChanged()
}

func (h *view) PageUpp() {
	_, y := h.Size()
	move := y - 2
	if h.firstLine-move < 0 {
		move = h.firstLine
	}
	if move < 1 {
		return
	}
	h.firstLine = h.firstLine - move
	h.lastLine = h.lastLine - move
	h.currentLine = h.currentLine - move
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
