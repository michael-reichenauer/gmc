package repoview

import (
	"github.com/jroimartin/gocui"
	"github.com/michael-reichenauer/gmc/repoview/viewmodel"
	"github.com/michael-reichenauer/gmc/utils/log"
	"github.com/michael-reichenauer/gmc/utils/ui"
)

type DiffView struct {
	ui.View
	vm             *diffVM
	mainController mainController
}

func NewDiffView(uiHandler *ui.UI, model *viewmodel.Service, mainController mainController) *DiffView {
	h := &DiffView{
		vm:             NewDiffVM(model),
		mainController: mainController,
	}
	h.View = uiHandler.NewView(h.viewData)
	h.Properties().OnLoad = h.onLoad
	h.View.Properties().Name = "DiffView"
	h.View.Properties().HasFrame = true
	h.View.Properties().Title = "Diff"
	return h
}

func (h *DiffView) onLoad() {
	h.SetKey(gocui.KeyEsc, gocui.ModNone, h.mainController.HideDiff)
	h.SetKey(gocui.KeyCtrlC, gocui.ModNone, h.mainController.HideDiff)
	h.SetKey(gocui.KeyCtrlC, gocui.ModNone, h.mainController.HideDiff)
	h.SetKey('q', gocui.ModNone, h.mainController.HideDiff)
	h.SetKey(gocui.KeyCtrlQ, gocui.ModNone, h.mainController.HideDiff)
	h.NotifyChanged()
}

func (h *DiffView) viewData(viewPort ui.ViewPage) ui.ViewData {
	diff, err := h.vm.getCommitDiff(viewPort)
	if err != nil {
		return ui.ViewData{}
	}
	log.Infof("view data")
	return ui.ViewData{Lines: diff.lines, FirstIndex: diff.firstIndex, Total: diff.total}
}

func (h *DiffView) SetIndex(index int) {
	h.vm.SetIndex(index)
}
