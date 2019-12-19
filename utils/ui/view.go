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
	SetCurrent()
	NotifyChanged()
	SetKey(key interface{}, modifier gocui.Modifier, handler func())
	CurrentLine() int
	Clear()
	PostOnUIThread(func())
}

type ViewHandler struct {
	gui     *gocui.Gui
	guiView *gocui.View

	properties  *Properties
	bounds      func(w, h int) Rect
	ViewName    string
	ViewData    ViewData
	FirstLine   int
	LastLine    int
	currentLine int
	maxX        int
	maxY        int
}

func newView(uiHandler *Handler) *ViewHandler {
	return &ViewHandler{
		gui:        uiHandler.Gui(),
		ViewName:   uiHandler.NewViewName(),
		properties: &Properties{}}
}

func (h *ViewHandler) Show(bounds Rect) {
	log.Infof("show before set")
	if gv, err := h.gui.SetView(h.ViewName, bounds.X-1, bounds.Y-1, bounds.W+1, bounds.H+1); err != nil {
		log.Infof("after set")
		if err != gocui.ErrUnknownView {
			log.Fatal(err)
		}
		log.Infof("after more")
		h.guiView = gv
		_, vy := h.guiView.Size()
		h.FirstLine = 0
		h.LastLine = vy - 1

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

func (h *ViewHandler) SetBounds(bounds Rect) {
	if _, err := h.gui.SetView(h.ViewName, bounds.X-1, bounds.Y-1, bounds.W+1, bounds.H+1); err != nil {
		log.Fatal(err)
	}
}

func (h *ViewHandler) SetCurrent() {
	if _, err := h.gui.SetCurrentView(h.ViewName); err != nil {
		log.Fatal(err)
	}
}

func (h ViewHandler) CurrentLine() int {
	return h.currentLine
}

func (h *ViewHandler) Properties() *Properties {
	return h.properties
}

func (h *ViewHandler) SetBound(bounds func(w, h int) Rect) {
	h.bounds = bounds
}

func (h *ViewHandler) PostOnUIThread(f func()) {
	h.gui.Update(func(g *gocui.Gui) error {
		f()
		return nil
	})
}

func (h *ViewHandler) NotifyChanged() {
	h.gui.Update(func(g *gocui.Gui) error {
		view, err := g.View(h.ViewName)
		if err != nil {
			return err
		}
		view.Clear()
		x, y := view.Size()
		h.LastLine = h.FirstLine + y - 1
		log.Infof("before %d %d %d %d", x, y, h.FirstLine, h.LastLine)
		h.ViewData = h.properties.OnViewData(ViewPort{Width: x, First: h.FirstLine, Last: h.LastLine, Current: h.currentLine})
		log.Infof("after %d %d %d %d", x, y, h.FirstLine, h.LastLine)
		h.FirstLine = h.ViewData.First
		h.LastLine = h.ViewData.Last
		h.currentLine = h.ViewData.Current
		_, _ = view.Write([]byte(h.ViewData.Text))
		return nil
	})
}

//func (h *ViewHandler) Resize(ww, wh int) {
//	b := h.getBounds(ww, wh)
//	_, err := h.gui.SetView(h.ViewName, b.X-1, b.Y-1, b.W+1, b.H+1)
//	if err != nil {
//		log.Fatalf("failed, %v", err)
//	}
//}

func (h *ViewHandler) Close() {
	if h.properties.OnClose != nil {
		h.properties.OnClose()
	}
	if err := h.gui.DeleteView(h.ViewName); err != nil {
		log.Fatal(err)
	}
}

func (h *ViewHandler) SetKey(key interface{}, modifier gocui.Modifier, handler func()) {
	if err := h.gui.SetKeybinding(
		h.ViewName, key, modifier,
		func(gui *gocui.Gui, view *gocui.View) error {
			handler()
			return nil
		}); err != nil {
		log.Fatalf("failed, %v", err)
	}
}

//func (h *ViewHandler) getBounds(ww, wh int) Rect {
//	if h.bounds == nil {
//		return Rect{X: 0, Y: 0, W: ww, H: wh}
//	}
//	return h.bounds(ww, wh)
//}

func (h *ViewHandler) Clear() {
	h.guiView.Clear()
}

func (h *ViewHandler) Cursor() (int, int) {
	return h.guiView.Cursor()
}

func (h *ViewHandler) SetCursor(x int, y int) error {
	return h.guiView.SetCursor(x, y)
}

func (h *ViewHandler) Size() (int, int) {
	return h.guiView.Size()
}

func (h *ViewHandler) CursorUp() {
	if h.currentLine <= 0 {
		return
	}

	cx, cy := h.Cursor()
	if err := h.SetCursor(cx, cy-1); err != nil {
		log.Fatal(err)
	}

	h.currentLine = h.currentLine - 1
	if h.currentLine < h.FirstLine {
		move := h.FirstLine - h.currentLine
		h.FirstLine = h.FirstLine - move
		h.LastLine = h.LastLine - move
	}
	h.NotifyChanged()
}

func (h *ViewHandler) CursorDown() {
	if h.currentLine >= h.ViewData.MaxLines-1 {
		return
	}
	cx, cy := h.Cursor()
	if err := h.SetCursor(cx, cy+1); err != nil {
		log.Fatal(err)
	}

	h.currentLine = h.currentLine + 1
	if h.currentLine > h.LastLine {
		move := h.currentLine - h.LastLine
		h.FirstLine = h.FirstLine + move
		h.LastLine = h.LastLine + move
	}
	h.NotifyChanged()
}
func (h *ViewHandler) PageDown() {
	_, y := h.Size()
	move := y - 2
	if h.LastLine+move >= h.ViewData.MaxLines-1 {
		move = h.ViewData.MaxLines - 1 - h.LastLine
	}
	if move < 1 {
		return
	}
	h.FirstLine = h.FirstLine + move
	h.LastLine = h.LastLine + move
	h.currentLine = h.currentLine + move
	h.NotifyChanged()
}

func (h *ViewHandler) PageUpp() {
	_, y := h.Size()
	move := y - 2
	if h.FirstLine-move < 0 {
		move = h.FirstLine
	}
	if move < 1 {
		return
	}
	h.FirstLine = h.FirstLine - move
	h.LastLine = h.LastLine - move
	h.currentLine = h.currentLine - move
	h.NotifyChanged()
}

//func (h *Handler) setCursor(gui *gocui.Gui, view *gocui.View, line int) error {
//	log.Infof("Set line %d", line)
//
//	if line >= h.viewHandler.ViewData.MaxLines {
//		return nil
//	}
//	cx, _ := view.Cursor()
//	_ = view.SetCursor(cx, line)
//
//	h.viewHandler.CurrentLine = line
//	if h.viewHandler.CurrentLine > h.viewHandler.LastLine {
//		move := h.viewHandler.CurrentLine - h.viewHandler.LastLine
//		h.viewHandler.FirstLine = h.viewHandler.FirstLine + move
//		h.viewHandler.LastLine = h.viewHandler.LastLine + move
//	}
//	h.viewHandler.NotifyChanged()
//
//	return nil
//}
