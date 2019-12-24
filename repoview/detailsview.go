package repoview

import (
	"github.com/michael-reichenauer/gmc/repoview/viewmodel"
	"github.com/michael-reichenauer/gmc/utils/ui"
)

type DetailsView struct {
	ui.View
	vm            *detailsVM
	selectedIndex int
}

func newDetailsView(uiHandler *ui.UI, model *viewmodel.Model) *DetailsView {
	h := &DetailsView{
		vm: newDetailsVM(model),
	}
	h.View = uiHandler.NewView(h.viewData)
	return h
}

func (h *DetailsView) viewData(viewPort ui.ViewPort) ui.ViewData {
	details, err := h.vm.getCommitDetails(viewPort, h.selectedIndex)
	if err != nil {
		return ui.ViewData{}
	}
	return ui.ViewData{Lines: details.lines}
}

func (h *DetailsView) SetCurrent(selectedIndex int) {
	h.selectedIndex = selectedIndex
	h.NotifyChanged()
}
