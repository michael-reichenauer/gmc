package repoview

import (
	"fmt"
	"github.com/jroimartin/gocui"
	"github.com/michael-reichenauer/gmc/utils"
	"github.com/michael-reichenauer/gmc/utils/log"
	"github.com/michael-reichenauer/gmc/utils/ui"
)

type Handler struct {
	uiHandler     *ui.Handler
	view          *ui.View
	vm            *repoVM
	repoPath      string
	isSelected    bool
	currentBranch string
}

func New(uiHandler *ui.Handler, repoPath string) *Handler {
	return &Handler{
		uiHandler: uiHandler,
		repoPath:  repoPath,
		vm:        newRepoVM(repoPath),
	}
}

func (h *Handler) Properties() ui.Properties {
	return ui.Properties{
		Title:         "",
		OnLoad:        h.OnLoad,
		IsCurrentView: true,
	}
}

func (h *Handler) GetViewData(width, firstLine, lastLine, selected int) ui.ViewData {
	repoPage, err := h.vm.GetRepoPage(width, firstLine, lastLine, selected)
	if err != nil {
		return ui.ViewData{Text: ui.Red(fmt.Sprintf("Error: %v", err)), MaxLines: 1}
	}
	if h.currentBranch != repoPage.currentBranchName || h.repoPath != repoPage.repoPath {
		h.setWindowTitle(repoPage.repoPath, repoPage.currentBranchName)
	}

	if !h.isSelected && repoPage.currentCommitIndex != -1 {
		h.isSelected = true
		h.SetCursor(repoPage.currentCommitIndex)
	}

	return ui.ViewData{Text: repoPage.text, MaxLines: repoPage.lines}
}

func (h *Handler) OnLoad(view *ui.View) {
	h.view = view
	h.vm.Load()
	h.setWindowTitle(h.repoPath, "")
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

func (h *Handler) onEnter() {
	log.Infof("Enter on commitVM %d", h.view.CurrentLine)
}

func (h *Handler) onRight() {
	h.vm.OpenBranch(h.view.CurrentLine)
	h.view.NotifyChanged()
}

func (h *Handler) onLeft() {
	h.vm.CloseBranch(h.view.CurrentLine)
	h.view.NotifyChanged()
}

func (h *Handler) onRefresh() {
	h.view.View.Clear()
	h.view.Gui.Update(func(g *gocui.Gui) error {
		h.vm.Refresh()
		h.view.NotifyChanged()
		return nil
	})

}

func (h *Handler) setWindowTitle(path, branch string) {
	_, _ = utils.SetConsoleTitle(fmt.Sprintf("gmc: %s - %s", path, branch))
}
func (h *Handler) SetCursor(line int) {

	//	h.setCursor(g, view, line)

}

func (h *Handler) cursorUp() {

	if h.view.CurrentLine <= 0 {
		return
	}

	cx, cy := h.view.View.Cursor()
	_ = h.view.View.SetCursor(cx, cy-1)

	h.view.CurrentLine = h.view.CurrentLine - 1
	if h.view.CurrentLine < h.view.FirstLine {
		move := h.view.FirstLine - h.view.CurrentLine
		h.view.FirstLine = h.view.FirstLine - move
		h.view.LastLine = h.view.LastLine - move
	}
	h.view.NotifyChanged()
}

func (h *Handler) cursorDown() {
	if h.view.CurrentLine >= h.view.ViewData.MaxLines-1 {
		return
	}
	cx, cy := h.view.View.Cursor()
	_ = h.view.View.SetCursor(cx, cy+1)

	h.view.CurrentLine = h.view.CurrentLine + 1
	if h.view.CurrentLine > h.view.LastLine {
		move := h.view.CurrentLine - h.view.LastLine
		h.view.FirstLine = h.view.FirstLine + move
		h.view.LastLine = h.view.LastLine + move
	}
	h.view.NotifyChanged()
}
func (h *Handler) pageDown() {

	_, y := h.view.View.Size()
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
func (h *Handler) pageUpp() {
	_, y := h.view.View.Size()
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

func (h *Handler) setCursor(gui *gocui.Gui, view *gocui.View, line int) error {
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
