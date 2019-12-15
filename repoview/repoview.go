package repoview

import (
	"fmt"
	"github.com/jroimartin/gocui"
	"github.com/michael-reichenauer/gmc/utils"
	"github.com/michael-reichenauer/gmc/utils/ui"
)

type RepoView struct {
	uiHandler          *ui.Handler
	viewHandler        *ui.ViewHandler
	vm                 *repoVM
	repoPath           string
	isSelected         bool
	currentBranch      string
	isShowCommitStatus bool
}

func NewRepoView(uiHandler *ui.Handler, repoPath string) *RepoView {
	return &RepoView{
		uiHandler: uiHandler,
		repoPath:  repoPath,
		vm:        newRepoVM(repoPath),
	}
}

func (h *RepoView) Properties() ui.Properties {
	return ui.Properties{
		Title:  "",
		OnLoad: h.OnLoad,
	}
}

func (h *RepoView) GetViewData(viewPort ui.ViewPort) ui.ViewData {
	repoPage, err := h.vm.GetRepoPage(viewPort.Width, viewPort.First, viewPort.Last, viewPort.Current)
	if err != nil {
		return ui.ViewData{Text: ui.Red(fmt.Sprintf("Error: %v", err)), MaxLines: 1}
	}

	h.setWindowTitle(repoPage.repoPath, repoPage.currentBranchName, repoPage.commitStatus)

	if !h.isSelected && repoPage.currentCommitIndex != -1 {
		h.isSelected = true
		h.SetCursor(repoPage.currentCommitIndex)
	}

	return ui.ViewData{
		Text:     repoPage.text,
		MaxLines: repoPage.lines,
		First:    repoPage.first,
		Last:     repoPage.last,
		Current:  repoPage.current,
	}
}

func (h *RepoView) OnLoad(view *ui.ViewHandler) {
	h.viewHandler = view
	h.vm.Load()
	h.setWindowTitle(h.repoPath, "", "")
	h.viewHandler.SetKey(gocui.KeyCtrl5, gocui.ModNone, h.onRefresh)
	h.viewHandler.SetKey(gocui.KeyF5, gocui.ModNone, h.onRefresh)
	h.viewHandler.SetKey(gocui.KeyEnter, gocui.ModNone, h.onEnter)
	h.viewHandler.SetKey(gocui.KeyArrowLeft, gocui.ModNone, h.onLeft)
	h.viewHandler.SetKey(gocui.KeyArrowRight, gocui.ModNone, h.onRight)
	h.viewHandler.SetKey(gocui.KeyArrowDown, gocui.ModNone, h.cursorDown)
	h.viewHandler.SetKey(gocui.KeySpace, gocui.ModNone, h.pageDown)
	h.viewHandler.SetKey(gocui.KeyPgdn, gocui.ModNone, h.pageDown)
	h.viewHandler.SetKey(gocui.KeyPgup, gocui.ModNone, h.pageUpp)
	h.viewHandler.SetKey(gocui.KeyArrowUp, gocui.ModNone, h.cursorUp)
	h.viewHandler.NotifyChanged()
}

func (h *RepoView) onEnter() {
	h.isShowCommitStatus = !h.isShowCommitStatus
	h.viewHandler.NotifyChanged()
}

func (h *RepoView) onRight() {
	h.vm.OpenBranch(h.viewHandler.CurrentLine)
	h.viewHandler.NotifyChanged()
}

func (h *RepoView) onLeft() {
	h.vm.CloseBranch(h.viewHandler.CurrentLine)
	h.viewHandler.NotifyChanged()
}

func (h *RepoView) onRefresh() {
	h.viewHandler.Clear()
	h.viewHandler.RunOnUI(func() {
		h.vm.Refresh()
		h.viewHandler.NotifyChanged()
	})
}

func (h *RepoView) setWindowTitle(path, branch, status string) {
	statusTxt := ""
	//if h.isShowCommitStatus {
	statusTxt = fmt.Sprintf("  %s", status)
	//}
	_, _ = utils.SetConsoleTitle(fmt.Sprintf("gmc: %s - %s%s", path, branch, statusTxt))
}
func (h *RepoView) SetCursor(line int) {

	//	h.setCursor(g, view, line)

}

func (h *RepoView) cursorUp() {
	if h.viewHandler.CurrentLine <= 0 {
		return
	}

	cx, cy := h.viewHandler.Cursor()
	_ = h.viewHandler.SetCursor(cx, cy-1)

	h.viewHandler.CurrentLine = h.viewHandler.CurrentLine - 1
	if h.viewHandler.CurrentLine < h.viewHandler.FirstLine {
		move := h.viewHandler.FirstLine - h.viewHandler.CurrentLine
		h.viewHandler.FirstLine = h.viewHandler.FirstLine - move
		h.viewHandler.LastLine = h.viewHandler.LastLine - move
	}
	h.viewHandler.NotifyChanged()
}

func (h *RepoView) cursorDown() {
	if h.viewHandler.CurrentLine >= h.viewHandler.ViewData.MaxLines-1 {
		return
	}
	cx, cy := h.viewHandler.Cursor()
	_ = h.viewHandler.SetCursor(cx, cy+1)

	h.viewHandler.CurrentLine = h.viewHandler.CurrentLine + 1
	if h.viewHandler.CurrentLine > h.viewHandler.LastLine {
		move := h.viewHandler.CurrentLine - h.viewHandler.LastLine
		h.viewHandler.FirstLine = h.viewHandler.FirstLine + move
		h.viewHandler.LastLine = h.viewHandler.LastLine + move
	}
	h.viewHandler.NotifyChanged()
}
func (h *RepoView) pageDown() {

	_, y := h.viewHandler.Size()
	move := y - 2
	if h.viewHandler.LastLine+move >= h.viewHandler.ViewData.MaxLines-1 {
		move = h.viewHandler.ViewData.MaxLines - 1 - h.viewHandler.LastLine
	}
	if move < 1 {
		return
	}
	h.viewHandler.FirstLine = h.viewHandler.FirstLine + move
	h.viewHandler.LastLine = h.viewHandler.LastLine + move
	h.viewHandler.CurrentLine = h.viewHandler.CurrentLine + move
	h.viewHandler.NotifyChanged()
}
func (h *RepoView) pageUpp() {
	_, y := h.viewHandler.Size()
	move := y - 2
	if h.viewHandler.FirstLine-move < 0 {
		move = h.viewHandler.FirstLine
	}
	if move < 1 {
		return
	}
	h.viewHandler.FirstLine = h.viewHandler.FirstLine - move
	h.viewHandler.LastLine = h.viewHandler.LastLine - move
	h.viewHandler.CurrentLine = h.viewHandler.CurrentLine - move
	h.viewHandler.NotifyChanged()
}

//func (h *Handler) setCursor(gui *gocui.Gui, view *gocui.View, line int) error {
//	log.Infof("Set line %d", line)
//
//	if line >= h.viewHandler.ViewData.MaxLines {
//		return nil
//	}
//	cx, _ := view.Cursor()
//	_ = view.SetCursor(cx, line)
//
//	h.viewHandler.CurrentLine = line
//	if h.viewHandler.CurrentLine > h.viewHandler.LastLine {
//		move := h.viewHandler.CurrentLine - h.viewHandler.LastLine
//		h.viewHandler.FirstLine = h.viewHandler.FirstLine + move
//		h.viewHandler.LastLine = h.viewHandler.LastLine + move
//	}
//	h.viewHandler.NotifyChanged()
//
//	return nil
//}
