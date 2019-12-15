package ui

import (
	"fmt"
	"github.com/jroimartin/gocui"
	"github.com/michael-reichenauer/gmc/utils"
	"github.com/michael-reichenauer/gmc/utils/log"
)

type Properties struct {
	Title    string
	HasFrame bool
	OnLoad   func(view *ViewHandler)
	OnClose  func()
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

type Rect struct {
	X, Y, W, H int
}

type View interface {
	Properties() Properties
	GetViewData(viewPort ViewPort) ViewData
}

type ViewHandler struct {
	gui     *gocui.Gui
	guiView *gocui.View

	properties  Properties
	ViewName    string
	viewModel   View
	ViewData    ViewData
	FirstLine   int
	LastLine    int
	CurrentLine int
	maxX        int
	maxY        int
}

func (h *ViewHandler) RunOnUI(f func()) {
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
		view.Cursor()
		x, y := view.Size()
		h.LastLine = h.FirstLine + y - 1
		h.ViewData = h.viewModel.GetViewData(ViewPort{Width: x, First: h.FirstLine, Last: h.LastLine, Current: h.CurrentLine})
		h.FirstLine = h.ViewData.First
		h.LastLine = h.ViewData.Last
		h.CurrentLine = h.ViewData.Current
		_, _ = view.Write([]byte(h.ViewData.Text))
		return nil
	})
}

func (h *ViewHandler) Resized() {
	h.gui.Update(func(g *gocui.Gui) error {
		view, err := g.View(h.ViewName)
		if err != nil {
			return err
		}
		view.Clear()
		maxX, maxY := g.Size()
		bounds := Rect{0, 0, maxX - 1, maxY}
		_, _ = g.SetView(h.ViewName, bounds.X-1, bounds.Y-1, bounds.W, bounds.H)
		return nil
	})
}

func (h *ViewHandler) Close() {
	h.gui.Update(func(g *gocui.Gui) error {
		if h.properties.OnClose != nil {
			h.properties.OnClose()
		}
		return h.gui.DeleteView(h.ViewName)
	})
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

func newView(gui *gocui.Gui, viewModel View) *ViewHandler {
	viewName := utils.RandomString(10)
	properties := viewModel.Properties()

	return &ViewHandler{gui: gui, ViewName: viewName, viewModel: viewModel, properties: properties}
}

func (h *ViewHandler) show() {
	h.gui.Update(func(g *gocui.Gui) error {
		maxX, maxY := g.Size()
		bounds := Rect{0, 0, maxX - 1, maxY}
		if gv, err := g.SetView(h.ViewName, bounds.X-1, bounds.Y-1, bounds.W, bounds.H); err != nil {
			if err != gocui.ErrUnknownView {
				return err
			}

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

			if len(h.gui.Views()) == 1 {
				if _, err := h.gui.SetCurrentView(h.ViewName); err != nil {
					return err
				}
			}
		}
		if h.properties.OnLoad != nil {
			h.properties.OnLoad(h)
		}
		return nil
	})
}

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
