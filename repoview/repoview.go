package repoview

import (
	"fmt"
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
		OnEnter:       h.OnEnter,
		OnRight:       h.OnRight,
		OnLeft:        h.OnLeft,
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
		h.view.SetCursor(repoPage.currentCommitIndex)
	}

	return ui.ViewData{Text: repoPage.text, MaxLines: repoPage.lines}
}

func (h *Handler) OnEnter(line int) {
	log.Infof("Enter on commitVM %d", line)
}

func (h *Handler) OnRight(line int) {
	h.vm.OpenBranch(line)
	h.view.NotifyChanged()
}
func (h *Handler) OnLeft(line int) {
	h.vm.CloseBranch(line)
	h.view.NotifyChanged()
}

func (h *Handler) OnLoad(view *ui.View) {
	log.Infof("onload")
	h.view = view
	h.vm.Load()
	h.setWindowTitle(h.repoPath, "")
	h.view.NotifyChanged()
}

func (h *Handler) setWindowTitle(path, branch string) {
	_, _ = utils.SetConsoleTitle(fmt.Sprintf("gmc: %s - %s", path, branch))
}
