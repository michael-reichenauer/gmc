package repoview

import (
	"github.com/jroimartin/gocui"
	"github.com/michael-reichenauer/gmc/repoview/viewmodel"
	"github.com/michael-reichenauer/gmc/utils/ui"
)

type DiffView struct {
	ui.View
	vm             *diffVM
	mainController mainController
	page           int
}

func NewDiffView(uiHandler *ui.UI, model *viewmodel.Service, mainController mainController) *DiffView {
	h := &DiffView{
		vm:             NewDiffVM(model),
		mainController: mainController,
	}
	h.View = uiHandler.NewViewFromPageFunc(h.viewData)
	h.Properties().OnLoad = h.onLoad
	h.View.Properties().Name = "DiffView"
	h.View.Properties().HasFrame = true
	h.View.Properties().Title = " Diff Unified "
	return h
}

func (h *DiffView) onLoad() {
	h.SetKey(gocui.KeyEsc, gocui.ModNone, h.mainController.HideDiff)
	h.SetKey(gocui.KeyCtrlC, gocui.ModNone, h.mainController.HideDiff)
	h.SetKey(gocui.KeyCtrlC, gocui.ModNone, h.mainController.HideDiff)
	h.SetKey(gocui.KeyArrowLeft, gocui.ModNone, h.onLeft)
	h.SetKey(gocui.KeyArrowRight, gocui.ModNone, h.onRight)
	h.SetKey('q', gocui.ModNone, h.mainController.HideDiff)
	h.SetKey(gocui.KeyCtrlQ, gocui.ModNone, h.mainController.HideDiff)
	h.NotifyChanged()
}

func (h *DiffView) viewData(viewPort ui.ViewPage) ui.ViewData {
	diff, err := h.vm.getCommitDiff(viewPort)
	if err != nil {
		return ui.ViewData{}
	}
	return ui.ViewData{Lines: diff.lines, FirstIndex: diff.firstIndex, Total: diff.total}
}

func (h *DiffView) SetIndex(index int) {
	h.page = 0
	h.SetTitle(" Diff Unified ")
	h.vm.SetIndex(index)
}

func (h *DiffView) onLeft() {
	if h.page >= 0 {
		h.page--
	} else {
		return
	}
	if h.page == 0 {
		h.SetTitle(" Diff Unified ")
	} else if h.page == -1 {
		h.SetTitle(" Diff Before ")
	}
	h.vm.SetLeft(h.page)
	h.NotifyChanged()
}

func (h *DiffView) onRight() {
	if h.page <= 0 {
		h.page++
	} else {
		return
	}
	if h.page == 0 {
		h.SetTitle(" Diff Unified ")
	} else if h.page == 1 {
		h.SetTitle(" Diff After ")
	}
	h.vm.SetRight(h.page)
	h.NotifyChanged()
}
