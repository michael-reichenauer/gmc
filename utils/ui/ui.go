package ui

import (
	"github.com/jroimartin/gocui"
	"github.com/michael-reichenauer/gmc/utils"
	"github.com/michael-reichenauer/gmc/utils/log"
	"github.com/nsf/termbox-go"
)

type Rect struct {
	X, Y, W, H int
}

type Notifier interface {
	NotifyChanged()
}

type Runner interface {
	PostOnUIThread(func())
}

type UI struct {
	gui            *gocui.Gui
	isInitialized  bool
	runFunc        func()
	maxX           int
	maxY           int
	OnResizeWindow func()
	currentView    *view
}

func NewUI() *UI {
	return &UI{}
}

func (h *UI) NewView(text string) View {
	return newView(h, text)
}

func (h *UI) NewMenu(title string) *Menu {
	return newMenu(h, title)
}

func (h *UI) NewViewFromPageFunc(viewData func(viewPort ViewPage) ViewPageData) View {
	return newViewFromPageFunc(h, viewData)
}

func (h *UI) NewViewFromTextFunc(viewText func(viewPort ViewPage) string) View {
	return newViewFromTextFunc(h, viewText)
}

func (h *UI) Run(runFunc func()) {
	h.runFunc = runFunc

	gui, err := gocui.NewGui(gocui.OutputNormal)
	if err != nil {
		panic(log.Fatal(err))
	}
	h.gui = gui
	defer gui.Close()

	gui.SetManagerFunc(h.layout)
	gui.InputEsc = true
	gui.BgColor = gocui.ColorBlack
	gui.Cursor = false
	gui.Mouse = true

	if err = gui.MainLoop(); err != nil && err != gocui.ErrQuit {
		panic(log.Fatal(err))
	}
}

func (h *UI) PostOnUIThread(f func()) {
	h.gui.Update(func(g *gocui.Gui) error {
		f()
		return nil
	})
}

func (h *UI) Gui() *gocui.Gui {
	return h.gui
}

func (h *UI) NewViewName() string {
	return utils.RandomString(10)
}

func (h *UI) WindowSize() (width, height int) {
	return h.gui.Size()
}

func (h *UI) SetKeyBinding(viewName string, key interface{}, mod gocui.Modifier, handler func(*gocui.Gui, *gocui.View) error) {
	if err := h.gui.SetKeybinding(viewName, key, mod, handler); err != nil {
		panic(log.Fatal(err))
	}
}

func SetWindowTitle(text string) {
	_, _ = utils.SetConsoleTitle(text)
}

func (h *UI) CurrentView() View {
	return h.currentView
}

func (h *UI) SetCurrentView(v View) {
	h.currentView = v.(*view)
	if h.currentView == nil {
		return
	}
	if _, err := h.gui.SetCurrentView(h.currentView.viewName); err != nil {
		panic(log.Fatal(err))
	}
}

func quit(g *gocui.Gui, v *gocui.View) error {
	log.Infof("Quiting")
	return gocui.ErrQuit
}

func (h *UI) layout(gui *gocui.Gui) error {
	// Resize window and notify all views if console window is resized
	maxX, maxY := gui.Size()
	if maxX != h.maxX || maxY != h.maxY {
		h.maxX = maxX
		h.maxY = maxY
		if h.OnResizeWindow != nil {
			h.OnResizeWindow()
		}
		if !h.gui.Cursor {
			termbox.SetCursor(0, 0) // workaround for hiding the cursor
		}
	}

	if h.isInitialized {
		return nil
	}
	h.isInitialized = true
	go h.runFunc()
	return nil
}

func (h *UI) Quit() {
	h.gui.Update(func(gui *gocui.Gui) error {
		return gocui.ErrQuit
	})
}
