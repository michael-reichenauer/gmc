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

type Handler struct {
	gui            *gocui.Gui
	isInitialized  bool
	runFunc        func()
	maxX           int
	maxY           int
	OnResizeWindow func()
}

func NewUI() *Handler {
	return &Handler{}
}
func (h *Handler) NewView() View {
	return newView(h)
}

//func (h *Handler) Show(view View, bounds func(w, h int) Rect) *ViewHandler {
//	viewHandler := newView(h.gui, view)
//	viewHandler.SetBound(bounds)
//	h.views = append(h.views, viewHandler)
//	viewHandler.show()
//	return viewHandler
//}

func (h *Handler) Run(runFunc func()) {
	h.runFunc = runFunc

	gui, err := gocui.NewGui(gocui.OutputNormal)
	if err != nil {
		log.Fatalf("failed, %v", err)
	}
	h.gui = gui
	defer gui.Close()

	gui.InputEsc = true
	// gui.Cursor = true
	//g.Mouse = true
	gui.BgColor = gocui.ColorBlack
	gui.Cursor = false
	gui.SetManagerFunc(h.layout)

	h.SetKeyBinding("", gocui.KeyCtrlC, gocui.ModNone, quit)
	h.SetKeyBinding("", 'q', gocui.ModNone, quit)
	h.SetKeyBinding("", gocui.KeyCtrlQ, gocui.ModNone, quit)
	if err = gui.MainLoop(); err != nil && err != gocui.ErrQuit {
		log.Fatalf("failed, %v", err)
	}
}

func (h *Handler) SetKeyBinding(viewName string, key interface{}, mod gocui.Modifier, handler func(*gocui.Gui, *gocui.View) error) {
	if err := h.gui.SetKeybinding(viewName, key, mod, handler); err != nil {
		log.Fatalf("failed, %v", err)
	}
}

func SetWindowTitle(text string) {
	_, _ = utils.SetConsoleTitle(text)
}

func quit(g *gocui.Gui, v *gocui.View) error {
	log.Infof("Quiting")
	return gocui.ErrQuit
}

func (h *Handler) layout(gui *gocui.Gui) error {
	maxX, maxY := gui.Size()
	log.Infof("layout %d %d %d %d", maxX, maxY, h.maxX, h.maxY)

	// Resize window and notify all views if console window is resized
	if maxX != h.maxX || maxY != h.maxY {
		log.Infof("resize %d %d", maxX, maxY)
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

func (h *Handler) Gui() *gocui.Gui {
	return h.gui
}

func (h *Handler) NewViewName() string {
	return utils.RandomString(10)
}

func (h *Handler) WindowSize() (width, height int) {
	return h.gui.Size()
}
