package ui

import (
	"github.com/jroimartin/gocui"
	"github.com/michael-reichenauer/gmc/utils"
	"github.com/michael-reichenauer/gmc/utils/log"
)

type Handler struct {
	gui           *gocui.Gui
	isInitialized bool
	runFunc       func()
	maxX          int
	maxY          int
	views         []*ViewHandler
}

func NewUI() *Handler {
	return &Handler{}
}

func (h *Handler) Show(viewModel View) *ViewHandler {
	view := newView(h.gui, viewModel)
	h.views = append(h.views, view)
	view.show()
	return view
}

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

	if err = gui.SetKeybinding("", gocui.KeyCtrlC, gocui.ModNone, quit); err != nil {
		log.Fatalf("failed, %v", err)
	}
	if err = gui.SetKeybinding("", gocui.KeyEsc, gocui.ModNone, quit); err != nil {
		log.Fatalf("failed, %v", err)
	}
	if err = gui.SetKeybinding("", 'q', gocui.ModNone, quit); err != nil {
		log.Fatalf("failed, %v", err)
	}
	if err = gui.SetKeybinding("", gocui.KeyCtrlQ, gocui.ModNone, quit); err != nil {
		log.Fatalf("failed, %v", err)
	}

	if err = gui.MainLoop(); err != nil && err != gocui.ErrQuit {
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

	// Resize all views if console window is resized
	if maxX != h.maxX || maxY != h.maxY {
		h.maxX = maxX
		h.maxY = maxY
		for _, v := range h.views {
			v.Resized()
			v.NotifyChanged()
		}
	}

	if h.isInitialized {
		return nil
	}
	h.isInitialized = true
	go h.runFunc()

	return nil
}
