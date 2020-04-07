package ui

import (
	"fmt"
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
	gui               *gocui.Gui
	isInitialized     bool
	runFunc           func()
	maxX              int
	maxY              int
	OnResizeWindow    func()
	currentViewsStack []*view
}

func NewUI() *UI {
	return &UI{}
}

func (h *UI) NewView(text string) View {
	return newView(h, viewDataFromText(text))
}

func (h *UI) NewMenu(title string) *Menu {
	return newMenu(h, title)
}

func (h *UI) NewViewFromPageFunc(viewData func(viewPort ViewPage) ViewText) View {
	return newView(h, viewData)
}

func (h *UI) NewViewFromTextFunc(viewText func(viewPort ViewPage) string) View {
	return newView(h, viewDataFromTextFunc(viewText))
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

func (h *UI) CenterBounds(maxWidth, maxHeight int) Rect {
	windowWidth, windowHeight := h.WindowSize()
	width := maxWidth
	height := maxHeight
	if maxWidth == 0 {
		width = windowWidth
	}
	if maxHeight == 0 {
		height = windowHeight
	}

	if width > windowWidth-4 {
		width = windowWidth - 4
	}
	if height > windowHeight-4 {
		height = windowHeight - 4
	}
	x := (windowWidth - width) / 2
	y := (windowHeight - height) / 2
	return Rect{X: x, Y: y, W: width, H: height}
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

func (h *UI) currentView() *view {
	if len(h.currentViewsStack) == 0 {
		return nil
	}
	return h.currentViewsStack[len(h.currentViewsStack)-1]
}

func (h *UI) setCurrentView(v *view) {
	previousCurrentView := h.currentView()
	if _, err := h.gui.SetCurrentView(v.viewName); err != nil {
		panic(log.Fatal(err))
	}
	log.Infof("Set current %q %q", v.viewName, v.properties.Name)
	h.addCurrentView(v)
	if previousCurrentView != nil {
		previousCurrentView.NotifyChanged()
	}
}

func (h *UI) addCurrentView(v *view) {
	h.removeCurrentView(v)
	h.currentViewsStack = append(h.currentViewsStack, v)
}

func (h *UI) removeCurrentView(v *view) {
	var views []*view
	for _, cv := range h.currentViewsStack {
		if cv == v {
			continue
		}
		views = append(views, cv)
	}
	h.currentViewsStack = views
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
		termbox.SetCursor(0, 0) // workaround for hiding the cursor
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

func (h *UI) ShowCursor(isShow bool) {
	h.gui.Cursor = isShow
}

func (h *UI) closeView(v *view) {
	if err := h.gui.DeleteView(v.viewName); err != nil {
		panic(log.Fatal(err))
	}

	isCurrent := h.currentView() == v
	h.removeCurrentView(v)
	if isCurrent {
		cv := h.currentView()
		if cv != nil {
			h.setCurrentView(cv)
			h.ShowCursor(cv.properties.IsEditable)
		}
	} else {
		h.removeCurrentView(v)
	}
}

func (h *UI) deleteKey(v *view, key interface{}) {
	if err := h.gui.DeleteKeybinding(v.viewName, key, gocui.ModNone); err != nil {
		panic(log.Fatal(err))
	}
}

func (h *UI) SetKey(v *view, key interface{}, handler func()) {
	if err := h.gui.SetKeybinding(v.viewName, key, gocui.ModNone, func(gui *gocui.Gui, view *gocui.View) error {
		handler()
		return nil
	}); err != nil {
		panic(log.Fatal(err))
	}
}

func (h *UI) setBottom(v *view) {
	if _, err := h.gui.SetViewOnBottom(v.viewName); err != nil {
		panic(log.Fatal(err))
	}
}

func (h *UI) setBounds(v *view, bounds Rect) {
	if _, err := h.gui.SetView(v.viewName, bounds.X, bounds.Y, bounds.W, bounds.H); err != nil {
		panic(log.Fatal(err))
	}
}

func (h *UI) setTop(v *view) {
	if _, err := h.gui.SetViewOnTop(v.viewName); err != nil {
		panic(log.Fatal(err))
	}
}

func (h *UI) createView(v *view, mb Rect) *gocui.View {
	if guiView, err := h.gui.SetView(v.viewName, mb.X, mb.Y, mb.W, mb.H); err != nil {
		if err != gocui.ErrUnknownView {
			panic(log.Fatalf(err, "%s %+v,%d,%d,%d", v.viewName, mb))
		}
		return guiView
	}
	panic(log.Fatalf(fmt.Errorf("view altready created"), "%s %+v,%d,%d,%d", v.viewName, mb))
}
