package cui

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

type UI interface {
	NewView(text string) View
	NewViewFromPageFunc(f func(viewPort ViewPage) ViewText) View
	NewViewFromTextFunc(f func(viewPage ViewPage) string) View
	Post(f func())
	ShowProgress(format string, v ...interface{}) Progress
	ShowMessageBox(title string, format string, v ...interface{})
	ShowErrorMessageBox(format string, v ...interface{})
	MessageBox(title, text string) *MessageBox
	ResizeAllViews()
	NewMenu(title string) Menu
	Quit()
}

type ui struct {
	gui               *gocui.Gui
	isInitialized     bool
	runFunc           func()
	windowWidth       int
	windowHeight      int
	currentViewsStack []*view
	shownViews        []*view
}

func NewCommandUI() *ui {
	return &ui{}
}

func (t *ui) MessageBox(title, text string) *MessageBox {
	return NewMessageBox(t, text, title)
}

func (t *ui) ShowMessageBox(title, format string, v ...interface{}) {
	text := fmt.Sprintf(format, v...)
	msgBox := NewMessageBox(t, text, title)
	msgBox.Show()
}

func (t *ui) ShowErrorMessageBox(format string, v ...interface{}) {
	text := Red(fmt.Sprintf(format, v...))
	msgBox := NewMessageBox(t, text, "Error !")
	msgBox.Show()
}

func (t *ui) ShowProgress(format string, v ...interface{}) Progress {
	return showProgress(t, format, v...)
}

func (t *ui) NewMenu(title string) Menu {
	return newMenu(t, title)
}

func (t *ui) NewView(text string) View {
	return newView(t, viewDataFromText(text))
}

func (t *ui) NewViewFromPageFunc(viewData func(viewPort ViewPage) ViewText) View {
	return newView(t, viewData)
}

func (t *ui) NewViewFromTextFunc(viewText func(viewPort ViewPage) string) View {
	return newView(t, viewDataFromTextFunc(viewText))
}

func (t *ui) Run(runFunc func()) {
	t.runFunc = runFunc

	gui, err := gocui.NewGui(gocui.OutputNormal)
	if err != nil {
		panic(log.Fatal(err))
	}
	t.gui = gui
	defer gui.Close()

	gui.SetManagerFunc(t.layout)
	gui.InputEsc = true
	gui.BgColor = gocui.ColorBlack
	gui.Cursor = false
	gui.Mouse = true

	if err = gui.MainLoop(); err != nil && err != gocui.ErrQuit {
		panic(log.Fatal(err))
	}
}

func (t *ui) Post(f func()) {
	t.gui.Update(func(g *gocui.Gui) error {
		f()
		return nil
	})
}

func (t *ui) WindowSize() (width, height int) {
	return t.gui.Size()
}

func SetWindowTitle(text string) {
	_, _ = utils.SetConsoleTitle(text)
}

func (t *ui) currentView() *view {
	if len(t.currentViewsStack) == 0 {
		return nil
	}
	return t.currentViewsStack[len(t.currentViewsStack)-1]
}

func (t *ui) setCurrentView(v *view) {
	previousCurrentView := t.currentView()
	if _, err := t.gui.SetCurrentView(v.guiView.Name()); err != nil {
		panic(log.Fatal(err))
	}
	t.gui.Cursor = v.properties.IsEditable
	log.Infof("Set current %q %q", v.guiView.Name(), v.properties.Name)
	t.addCurrentView(v)
	if previousCurrentView != nil {
		previousCurrentView.NotifyChanged()
	}
}

func (t *ui) addCurrentView(v *view) {
	t.removeCurrentView(v)
	t.currentViewsStack = append(t.currentViewsStack, v)
}

func (t *ui) removeCurrentView(v *view) {
	var views []*view
	for _, cv := range t.currentViewsStack {
		if cv == v {
			continue
		}
		views = append(views, cv)
	}
	t.currentViewsStack = views
}

func (t *ui) layout(gui *gocui.Gui) error {
	// Resize window and notify all views if console window is resized
	windowWidth, windowHeight := gui.Size()
	if windowWidth != t.windowWidth || windowHeight != t.windowHeight {
		t.windowWidth = windowWidth
		t.windowHeight = windowHeight
		t.ResizeAllViews()
		termbox.SetCursor(0, 0) // workaround for hiding the cursor
	}

	if t.isInitialized {
		return nil
	}
	t.isInitialized = true
	go t.runFunc()
	return nil
}

func (t *ui) ResizeAllViews() {
	for _, v := range t.shownViews {
		v.resize(t.windowWidth, t.windowHeight)
		v.NotifyChanged()
	}
}

func (t *ui) CenterBounds(minWidth, minHeight, maxWidth, maxHeight int) Rect {
	bf := CenterBounds(minWidth, minHeight, maxWidth, maxHeight)
	ww, wh := t.WindowSize()
	return bf(ww, wh)
}

func (t *ui) Quit() {
	t.gui.Update(func(gui *gocui.Gui) error {
		return gocui.ErrQuit
	})
}

func (t *ui) showCursor(isShow bool) {
	t.gui.Cursor = isShow
}

func (t *ui) createView() *gocui.View {
	mb := Rect{-2, -2, -1, -1}
	name := utils.RandomString(10)
	if guiView, err := t.gui.SetView(name, mb.X, mb.Y, mb.W, mb.H); err != nil {
		if err != gocui.ErrUnknownView {
			panic(log.Fatalf(err, "%s %+v", name, mb))
		}
		return guiView
	}
	panic(log.Fatalf(fmt.Errorf("view already created"), "%s %+v", name, mb))
}

func (t *ui) setBounds(v *gocui.View, bounds Rect) {
	if v == nil {
		return
	}
	if _, err := t.gui.SetView(v.Name(), bounds.X, bounds.Y, bounds.W, bounds.H); err != nil {
		panic(log.Fatal(err))
	}
}

func (t *ui) closeView(v *view) {
	t.deleteView(v.guiView)

	isCurrent := t.currentView() == v
	t.removeCurrentView(v)
	if isCurrent {
		cv := t.currentView()
		if cv != nil {
			t.setCurrentView(cv)
			t.showCursor(cv.properties.IsEditable)
		}
	}
	t.removeShownView(v)
}

func (t *ui) deleteView(v *gocui.View) {
	if v == nil {
		return
	}
	if err := t.gui.DeleteView(v.Name()); err != nil {
		panic(log.Fatal(err))
	}
}

func (t *ui) deleteKey(v *gocui.View, key interface{}) {
	if err := t.gui.DeleteKeybinding(v.Name(), key, gocui.ModNone); err != nil {
		panic(log.Fatal(err))
	}
}

func (t *ui) setKey(v *gocui.View, key interface{}, handler func()) {
	if err := t.gui.SetKeybinding(v.Name(), key, gocui.ModNone, func(gui *gocui.Gui, view *gocui.View) error {
		handler()
		return nil
	}); err != nil {
		panic(log.Fatal(err))
	}
}

func (t *ui) setTop(v *gocui.View) {
	if v == nil {
		return
	}

	if _, err := t.gui.SetViewOnTop(v.Name()); err != nil {
		panic(log.Fatalf(err, "failed for %q", v.Name()))
	}
}

func (t *ui) setBottom(v *gocui.View) {
	if v == nil {
		return
	}
	if _, err := t.gui.SetViewOnBottom(v.Name()); err != nil {
		panic(log.Fatalf(err, "failed for %q", v.Name()))
	}
}

func (t *ui) addShownView(v *view) {
	t.removeShownView(v)
	t.shownViews = append(t.shownViews, v)
}

func (t *ui) removeShownView(v *view) {
	var views []*view
	for _, sv := range t.shownViews {
		if sv == v {
			continue
		}
		views = append(views, sv)
	}
	t.shownViews = views
}
