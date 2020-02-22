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

func NewDetailsView(uiHandler *ui.UI, model *viewmodel.Service) *DetailsView {
	h := &DetailsView{
		vm: NewDetailsVM(model),
	}
	h.View = uiHandler.NewViewFromTextFunc(h.viewData)
	h.View.Properties().Name = "DetailsView"
	return h
}

func (h *DetailsView) viewData(viewPage ui.ViewPage) string {
	details, err := h.vm.getCommitDetails(viewPage, h.currentIndex)
	if err != nil {
		return ""
	}
	return details
}

func (h *DetailsView) SetCurrent(selectedIndex int) {
	h.currentIndex = selectedIndex
	h.NotifyChanged()
}
