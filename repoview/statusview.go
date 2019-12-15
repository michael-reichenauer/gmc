package repoview

import (
	"github.com/jroimartin/gocui"
	"github.com/michael-reichenauer/gmc/utils/log"
	"github.com/michael-reichenauer/gmc/utils/ui"
)

type StatusViewHandler struct {
	uiHandler          *ui.Handler
	view               *ui.ViewHandler
	vm                 *statusVM
	repoPath           string
	isSelected         bool
	currentBranch      string
	isShowCommitStatus bool
}

func NewStatusView(uiHandler *ui.Handler, repoPath string) *StatusViewHandler {
	return &StatusViewHandler{
		uiHandler: uiHandler,
		repoPath:  repoPath,
		vm:        newStatusVM(),
	}
}

func (h *StatusViewHandler) Properties() ui.Properties {
	return ui.Properties{
		Title:         "",
		OnLoad:        h.OnLoad,
		IsCurrentView: true,
	}
}

func (h *StatusViewHandler) GetViewData(width, firstLine, lastLine, selected int) ui.ViewData {
	//repoPage, err := h.vm.GetRepoPage(width, firstLine, lastLine, selected)
	//if err != nil {
	//	return ui.ViewData{Text: ui.Red(fmt.Sprintf("Error: %v", err)), MaxLines: 1}
	//}
	//
	//
	//if !h.isSelected && repoPage.currentCommitIndex != -1 {
	//	h.isSelected = true
	//	h.SetCursor(repoPage.currentCommitIndex)
	//}

	return ui.ViewData{
		Text:     "status text",
		MaxLines: 1,
		First:    0,
		Last:     0,
		Current:  0,
	}
}

func (h *StatusViewHandler) OnLoad(view *ui.ViewHandler) {
	h.view = view
	//h.vm.Load()
	h.view.SetKey(gocui.KeyCtrl5, gocui.ModNone, h.onRefresh)
	h.view.SetKey(gocui.KeyF5, gocui.ModNone, h.onRefresh)
	h.view.SetKey(gocui.KeyEnter, gocui.ModNone, h.onEnter)
	h.view.SetKey(gocui.KeyArrowLeft, gocui.ModNone, h.onLeft)
	h.view.SetKey(gocui.KeyArrowRight, gocui.ModNone, h.onRight)
	h.view.SetKey(gocui.KeyArrowDown, gocui.ModNone, h.cursorDown)
	h.view.SetKey(gocui.KeySpace, gocui.ModNone, h.pageDown)
	h.view.SetKey(gocui.KeyPgdn, gocui.ModNone, h.pageDown)
	h.view.SetKey(gocui.KeyPgup, gocui.ModNone, h.pageUpp)
	h.view.SetKey(gocui.KeyArrowUp, gocui.ModNone, h.cursorUp)
	h.view.NotifyChanged()
}

func (h *StatusViewHandler) onEnter() {
	h.view.NotifyChanged()
}

func (h *StatusViewHandler) onRight() {
	h.view.NotifyChanged()
}

func (h *StatusViewHandler) onLeft() {
	h.view.NotifyChanged()
}

func (h *StatusViewHandler) onRefresh() {
	h.view.Clear()
	h.view.RunOnUI(func() {
		//h.vm.Refresh()
		h.view.NotifyChanged()
	})
}

func (h *StatusViewHandler) SetCursor(line int) {

	//	h.setCursor(g, view, line)

}

func (h *StatusViewHandler) cursorUp() {
	if h.view.CurrentLine <= 0 {
		return
	}

	cx, cy := h.view.Cursor()
	_ = h.view.SetCursor(cx, cy-1)

	h.view.CurrentLine = h.view.CurrentLine - 1
	if h.view.CurrentLine < h.view.FirstLine {
		move := h.view.FirstLine - h.view.CurrentLine
		h.view.FirstLine = h.view.FirstLine - move
		h.view.LastLine = h.view.LastLine - move
	}
	h.view.NotifyChanged()
}

func (h *StatusViewHandler) cursorDown() {
	if h.view.CurrentLine >= h.view.ViewData.MaxLines-1 {
		return
	}
	cx, cy := h.view.Cursor()
	_ = h.view.SetCursor(cx, cy+1)

	h.view.CurrentLine = h.view.CurrentLine + 1
	if h.view.CurrentLine > h.view.LastLine {
		move := h.view.CurrentLine - h.view.LastLine
		h.view.FirstLine = h.view.FirstLine + move
		h.view.LastLine = h.view.LastLine + move
	}
	h.view.NotifyChanged()
}
func (h *StatusViewHandler) pageDown() {
	_, y := h.view.Size()
	move := y - 2
	if h.view.LastLine+move >= h.view.ViewData.MaxLines-1 {
		move = h.view.ViewData.MaxLines - 1 - h.view.LastLine
	}
	if move < 1 {
		return
	}
	h.view.FirstLine = h.view.FirstLine + move
	h.view.LastLine = h.view.LastLine + move
	h.view.CurrentLine = h.view.CurrentLine + move
	h.view.NotifyChanged()
}
func (h *StatusViewHandler) pageUpp() {
	_, y := h.view.Size()
	move := y - 2
	if h.view.FirstLine-move < 0 {
		move = h.view.FirstLine
	}
	if move < 1 {
		return
	}
	h.view.FirstLine = h.view.FirstLine - move
	h.view.LastLine = h.view.LastLine - move
	h.view.CurrentLine = h.view.CurrentLine - move
	h.view.NotifyChanged()
}

func (h *StatusViewHandler) setCursor(gui *gocui.Gui, view *gocui.View, line int) error {
	log.Infof("Set line %d", line)

	if line >= h.view.ViewData.MaxLines {
		return nil
	}
	cx, _ := view.Cursor()
	_ = view.SetCursor(cx, line)

	h.view.CurrentLine = line
	if h.view.CurrentLine > h.view.LastLine {
		move := h.view.CurrentLine - h.view.LastLine
		h.view.FirstLine = h.view.FirstLine + move
		h.view.LastLine = h.view.LastLine + move
	}
	h.view.NotifyChanged()

	return nil
}
