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
	windowWidth       int
	windowHeight      int
	currentViewsStack []*view
	shownViews        []*view
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

func (h *UI) WindowSize() (width, height int) {
	return h.gui.Size()
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
	if _, err := h.gui.SetCurrentView(v.guiView.Name()); err != nil {
		panic(log.Fatal(err))
	}
	h.gui.Cursor = v.properties.IsEditable
	log.Infof("Set current %q %q", v.guiView.Name(), v.properties.Name)
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
	windowWidth, windowHeight := gui.Size()
	if windowWidth != h.windowWidth || windowHeight != h.windowHeight {
		h.windowWidth = windowWidth
		h.windowHeight = windowHeight
		h.ResizeAllViews()
		termbox.SetCursor(0, 0) // workaround for hiding the cursor
	}

	if h.isInitialized {
		return nil
	}
	h.isInitialized = true
	go h.runFunc()
	return nil
}

func (h *UI) ResizeAllViews() {
	for _, v := range h.shownViews {
		v.resize(h.windowWidth, h.windowHeight)
		v.NotifyChanged()
	}
}

func (h *UI) CenterBounds(minWidth, minHeight, maxWidth, maxHeight int) Rect {
	bf := CenterBounds(minWidth, minHeight, maxWidth, maxHeight)
	ww, wh := h.WindowSize()
	return bf(ww, wh)
}

func (h *UI) Quit() {
	h.gui.Update(func(gui *gocui.Gui) error {
		return gocui.ErrQuit
	})
}

func (h *UI) ShowCursor(isShow bool) {
	h.gui.Cursor = isShow
}

func (h *UI) createView() *gocui.View {
	mb := Rect{-2, -2, -1, -1}
	name := utils.RandomString(10)
	if guiView, err := h.gui.SetView(name, mb.X, mb.Y, mb.W, mb.H); err != nil {
		if err != gocui.ErrUnknownView {
			panic(log.Fatalf(err, "%s %+v,%d,%d,%d", name, mb))
		}
		return guiView
	}
	panic(log.Fatalf(fmt.Errorf("view already created"), "%s %+v,%d,%d,%d", name, mb))
}

func (h *UI) setBounds(v *gocui.View, bounds Rect) {
	if v == nil {
		return
	}
	if _, err := h.gui.SetView(v.Name(), bounds.X, bounds.Y, bounds.W, bounds.H); err != nil {
		panic(log.Fatal(err))
	}
}

func (h *UI) closeView(v *view) {
	h.deleteView(v.guiView)

	isCurrent := h.currentView() == v
	h.removeCurrentView(v)
	if isCurrent {
		cv := h.currentView()
		if cv != nil {
			h.setCurrentView(cv)
			h.ShowCursor(cv.properties.IsEditable)
		}
	}
	h.removeShownView(v)
}

func (h *UI) deleteView(v *gocui.View) {
	if v == nil {
		return
	}
	if err := h.gui.DeleteView(v.Name()); err != nil {
		panic(log.Fatal(err))
	}
}

func (h *UI) deleteKey(v *gocui.View, key interface{}) {
	if err := h.gui.DeleteKeybinding(v.Name(), key, gocui.ModNone); err != nil {
		panic(log.Fatal(err))
	}
}

func (h *UI) setKey(v *gocui.View, key interface{}, handler func()) {
	if err := h.gui.SetKeybinding(v.Name(), key, gocui.ModNone, func(gui *gocui.Gui, view *gocui.View) error {
		handler()
		return nil
	}); err != nil {
		panic(log.Fatal(err))
	}
}

func (h *UI) setTop(v *gocui.View) {
	if v == nil {
		return
	}

	if _, err := h.gui.SetViewOnTop(v.Name()); err != nil {
		panic(log.Fatalf(err, "failed for %q", v.Name()))
	}
}

func (h *UI) setBottom(v *gocui.View) {
	if v == nil {
		return
	}
	if _, err := h.gui.SetViewOnBottom(v.Name()); err != nil {
		panic(log.Fatalf(err, "failed for %q", v.Name()))
	}
}

func (h *UI) addShownView(v *view) {
	h.removeShownView(v)
	h.shownViews = append(h.shownViews, v)
}

func (h *UI) removeShownView(v *view) {
	var views []*view
	for _, sv := range h.shownViews {
		if sv == v {
			continue
		}
		views = append(views, sv)
	}
	h.shownViews = views
}
