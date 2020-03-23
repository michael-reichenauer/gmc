package repoview

import (
	"github.com/jroimartin/gocui"
	"github.com/michael-reichenauer/gmc/utils/ui"
)

type DiffUniView struct {
	ui.View
	vm          *diffUniVM
	mainService mainService
}

func NewDiffUniView(uiHandler *ui.UI,
	mainService mainService,
	diffGetter DiffGetter,
	commitID string,
) *DiffUniView {
	h := &DiffUniView{
		mainService: mainService,
	}
	h.vm = NewDiffUniVM(h, diffGetter, commitID)
	h.View = uiHandler.NewViewFromPageFunc(h.viewData)
	h.Properties().OnLoad = h.onLoad
	h.View.Properties().Name = "DiffView"
	h.View.Properties().HasFrame = true
	h.View.Properties().Title = " Diff Unified "
	return h
}

func (h *DiffUniView) onLoad() {
	h.SetKey(gocui.KeyEsc, gocui.ModNone, h.mainService.HideDiff)
	h.SetKey(gocui.KeyCtrlC, gocui.ModNone, h.mainService.HideDiff)
	h.SetKey(gocui.KeyCtrlC, gocui.ModNone, h.mainService.HideDiff)
	h.SetKey(gocui.KeyCtrlQ, gocui.ModNone, h.mainService.HideDiff)
	h.SetKey('q', gocui.ModNone, h.mainService.HideDiff)
	h.SetKey(gocui.KeyArrowLeft, gocui.ModNone, h.vm.onLeft)
	h.SetKey(gocui.KeyArrowRight, gocui.ModNone, h.vm.onRight)

	h.vm.load()
}

func (h *DiffUniView) viewData(viewPort ui.ViewPage) ui.ViewPageData {
	diff, err := h.vm.getCommitDiff(viewPort)
	if err != nil {
		return ui.ViewPageData{}
	}
	return ui.ViewPageData{Lines: diff.lines, FirstIndex: diff.firstIndex, Total: diff.total}
}
