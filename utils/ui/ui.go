package ui

import (
	"github.com/jroimartin/gocui"
	"gmc/utils/log"
)

type Handler struct {
	gui           *gocui.Gui
	isInitialized bool
	runFunc       func()
}

func NewUI() *Handler {
	return &Handler{}
}

func (h *Handler) Show(viewModel ViewModel) *View {
	view := newView(h.gui, viewModel)
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

	gui.SetManagerFunc(h.layout)

	if err = gui.SetKeybinding("", gocui.KeyCtrlC, gocui.ModNone, quit); err != nil {
		log.Fatalf("failed, %v", err)
	}
	if err = gui.SetKeybinding("", gocui.KeyEsc, gocui.ModNone, quit); err != nil {
		log.Fatalf("failed, %v", err)
	}
	if err = gui.SetKeybinding("", gocui.KeyBackspace, gocui.ModNone, quit); err != nil {
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

func quit(g *gocui.Gui, v *gocui.View) error {
	return gocui.ErrQuit
}

func (h *Handler) layout(gui *gocui.Gui) error {
	//maxX, maxY := gui.Size()
	//_, err := gui.SetView("main", 0, 0, maxX-1, maxY-1)
	//if err != gocui.ErrUnknownView {
	//	return err
	//}
	if h.isInitialized {
		return nil
	}
	h.isInitialized = true
	go h.runFunc()

	return nil
}
