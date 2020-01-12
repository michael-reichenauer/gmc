package repoview

import (
	"github.com/michael-reichenauer/gmc/repoview/viewmodel"
	"github.com/michael-reichenauer/gmc/utils/ui"
)

type DetailsView struct {
	ui.View
	vm           *detailsVM
	currentIndex int
}

func newDetailsView(uiHandler *ui.UI, model *viewmodel.Service) *DetailsView {
	h := &DetailsView{
		vm: newDetailsVM(model),
	}
	h.View = uiHandler.NewView(h.viewData)
	h.View.Properties().Name = "DetailsView"
	return h
}

func (h *DetailsView) viewData(viewPort ui.ViewPage) ui.ViewData {
	details, err := h.vm.getCommitDetails(viewPort, h.currentIndex)
	if err != nil {
		return ui.ViewData{}
	}
	return ui.ViewData{Lines: details.lines}
}

func (h *DetailsView) SetCurrent(selectedIndex int) {
	h.currentIndex = selectedIndex
	h.NotifyChanged()
}
