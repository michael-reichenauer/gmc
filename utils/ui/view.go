package ui

import (
	"fmt"
	"github.com/jroimartin/gocui"
	uuid "github.com/satori/go.uuid"
	"gmc/utils/log"
)

type Properties struct {
	Title         string
	HasFrame      bool
	Bounds        func(screenRect Rect) Rect
	IsCurrentView bool
	OnLoad        func(view *View)
	OnClose       func()
	OnEnter       func(currentLine int)
	OnLeft        func(currentLine int)
	OnRight       func(currentLine int)
}

type ViewData struct {
	Text     string
	MaxLines int
}

type Rect struct {
	X, Y, W, H int
}

type ViewModel interface {
	Properties() Properties
	GetViewData(width, firstLine, lastLine, currentLine int) ViewData
}

type View struct {
	Gui  *gocui.Gui
	View *gocui.View

	properties  Properties
	viewName    string
	viewModel   ViewModel
	viewData    ViewData
	firstLine   int
	lastLine    int
	currentLine int
}

func (h *View) NotifyChanged() {
	h.Gui.Update(func(g *gocui.Gui) error {
		view, err := g.View(h.viewName)
		if err != nil {
			return err
		}
		view.Clear()
		//maxX, maxY := g.Size()
		//bounds := h.properties.Bounds(Rect{0, 0, maxX, maxY})
		//view, _ = g.SetView(h.viewName, bounds.X-1, bounds.Y-1, bounds.W, bounds.H)
		x, _ := view.Size()
		h.viewData = h.viewModel.GetViewData(x, h.firstLine, h.lastLine, h.currentLine)
		_, _ = view.Write([]byte(h.viewData.Text))
		return nil
	})
}

func (h *View) SetCursor(line int) {
	h.Gui.Update(func(g *gocui.Gui) error {
		view, err := g.View(h.viewName)
		if err != nil {
			return err
		}

		h.setCursor(g, view, line)
		return nil
	})
}

func (h *View) Close() {
	h.Gui.Update(func(g *gocui.Gui) error {
		if h.properties.OnClose != nil {
			h.properties.OnClose()
		}
		return h.Gui.DeleteView(h.viewName)
	})
}

func newView(gui *gocui.Gui, viewModel ViewModel) *View {
	viewName := uuid.NewV4().String()
	//	viewName := "main"
	properties := viewModel.Properties()
	if properties.Bounds == nil {
		properties.Bounds = func(sr Rect) Rect { return Rect{0, 0, sr.W - 1, sr.H} }
	}

	return &View{Gui: gui, viewName: viewName, viewModel: viewModel, properties: properties}
}

func (h *View) show() {
	h.Gui.Update(func(g *gocui.Gui) error {
		maxX, maxY := g.Size()
		bounds := h.properties.Bounds(Rect{0, 0, maxX, maxY})
		if gv, err := g.SetView(h.viewName, bounds.X-1, bounds.Y-1, bounds.W, bounds.H); err != nil {
			if err != gocui.ErrUnknownView {
				return err
			}

			h.View = gv
			_, vy := h.View.Size()
			h.firstLine = 0
			h.lastLine = vy - 1

			h.View.Frame = h.properties.Title != "" || h.properties.HasFrame
			h.View.Editable = false
			h.View.Wrap = false
			h.View.Highlight = false
			h.View.SelBgColor = gocui.ColorBlue
			if h.properties.Title != "" {
				h.View.Title = fmt.Sprintf(" %s ", h.properties.Title)
			}
			if err := h.Gui.SetKeybinding(h.viewName, gocui.KeyArrowDown, gocui.ModNone, h.cursorDown); err != nil {
				log.Fatalf("failed, %v", err)
			}
			if err := h.Gui.SetKeybinding(h.viewName, gocui.KeySpace, gocui.ModNone, h.pageDown); err != nil {
				log.Fatalf("failed, %v", err)
			}
			if err := h.Gui.SetKeybinding(h.viewName, gocui.KeyPgdn, gocui.ModNone, h.pageDown); err != nil {
				log.Fatalf("failed, %v", err)
			}
			if err := h.Gui.SetKeybinding(h.viewName, gocui.KeyPgup, gocui.ModNone, h.pageUpp); err != nil {
				log.Fatalf("failed, %v", err)
			}
			if err := h.Gui.SetKeybinding(h.viewName, gocui.KeyArrowUp, gocui.ModNone, h.cursorUp); err != nil {
				log.Fatalf("failed, %v", err)
			}
			if err := h.Gui.SetKeybinding(h.viewName, gocui.KeyEnter, gocui.ModNone, h.enter); err != nil {
				log.Fatalf("failed, %v", err)
			}
			if err := h.Gui.SetKeybinding(h.viewName, gocui.KeyArrowLeft, gocui.ModNone, h.left); err != nil {
				log.Fatalf("failed, %v", err)
			}
			if err := h.Gui.SetKeybinding(h.viewName, gocui.KeyArrowRight, gocui.ModNone, h.right); err != nil {
				log.Fatalf("failed, %v", err)
			}
			if err := h.Gui.SetKeybinding(h.viewName, 'q', gocui.ModNone, h.escape); err != nil {
				log.Fatalf("failed, %v", err)
			}
			if h.properties.IsCurrentView {
				if _, err := h.Gui.SetCurrentView(h.viewName); err != nil {
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

// return gocui.ErrQuit
func (h *View) enter(gui *gocui.Gui, view *gocui.View) error {
	if view != nil && h.properties.OnEnter != nil {
		h.properties.OnEnter(h.currentLine)
	}
	return nil
}
func (h *View) left(gui *gocui.Gui, view *gocui.View) error {
	if view != nil && h.properties.OnLeft != nil {
		h.properties.OnLeft(h.currentLine)
	}
	return nil
}
func (h *View) right(gui *gocui.Gui, view *gocui.View) error {
	if view != nil && h.properties.OnRight != nil {
		h.properties.OnRight(h.currentLine)
	}
	return nil
}
func (h *View) escape(gui *gocui.Gui, view *gocui.View) error {
	return gocui.ErrQuit
}

func (h *View) cursorUp(gui *gocui.Gui, view *gocui.View) error {
	if view != nil {
		if h.currentLine <= 0 {
			return nil
		}

		cx, cy := view.Cursor()
		_ = view.SetCursor(cx, cy-1)

		h.currentLine = h.currentLine - 1
		if h.currentLine < h.firstLine {
			move := h.firstLine - h.currentLine
			h.firstLine = h.firstLine - move
			h.lastLine = h.lastLine - move
		}
		h.NotifyChanged()
	}
	return nil
}

func (h *View) cursorDown(gui *gocui.Gui, view *gocui.View) error {
	if view != nil {
		if h.currentLine >= h.viewData.MaxLines-1 {
			return nil
		}
		cx, cy := view.Cursor()
		_ = view.SetCursor(cx, cy+1)

		h.currentLine = h.currentLine + 1
		if h.currentLine > h.lastLine {
			move := h.currentLine - h.lastLine
			h.firstLine = h.firstLine + move
			h.lastLine = h.lastLine + move
		}
		h.NotifyChanged()
	}
	return nil
}
func (h *View) pageDown(gui *gocui.Gui, view *gocui.View) error {
	if view != nil {
		_, y := view.Size()
		move := y - 2
		if h.lastLine+move >= h.viewData.MaxLines-1 {
			move = h.viewData.MaxLines - 1 - h.lastLine
		}
		if move < 1 {
			return nil
		}
		h.firstLine = h.firstLine + move
		h.lastLine = h.lastLine + move
		h.currentLine = h.currentLine + move
		h.NotifyChanged()
	}
	return nil
}
func (h *View) pageUpp(gui *gocui.Gui, view *gocui.View) error {
	if view != nil {
		_, y := view.Size()
		move := y - 2
		if h.firstLine-move < 0 {
			move = h.firstLine
		}
		if move < 1 {
			return nil
		}
		h.firstLine = h.firstLine - move
		h.lastLine = h.lastLine - move
		h.currentLine = h.currentLine - move
		h.NotifyChanged()
	}
	return nil
}

func (h *View) setCursor(gui *gocui.Gui, view *gocui.View, line int) error {
	log.Infof("Set line %d", line)
	if view != nil {
		if line >= h.viewData.MaxLines {
			return nil
		}
		cx, _ := view.Cursor()
		_ = view.SetCursor(cx, line)

		h.currentLine = line
		if h.currentLine > h.lastLine {
			move := h.currentLine - h.lastLine
			h.firstLine = h.firstLine + move
			h.lastLine = h.lastLine + move
		}
		h.NotifyChanged()
	}
	return nil
}
