package repoview

import (
	"github.com/michael-reichenauer/gmc/repoview/viewmodel"
	"github.com/michael-reichenauer/gmc/utils/ui"
)

type DetailsView struct {
	ui.View
	vm *detailsVM
}

func NewDetailsView(uiHandler *ui.UI) *DetailsView {
	h := &DetailsView{vm: NewDetailsVM()}
	h.View = uiHandler.NewViewFromTextFunc(h.viewData)
	h.View.Properties().Name = "DetailsView"
	return h
}

func (h *DetailsView) viewData(viewPage ui.ViewPage) string {
	details, err := h.vm.getCommitDetails(viewPage)
	if err != nil {
		return ""
	}
	return details
}

func (h *DetailsView) SetCurrent(commit viewmodel.Commit) {
	h.vm.setCurrentCommit(commit)
	h.NotifyChanged()
}
