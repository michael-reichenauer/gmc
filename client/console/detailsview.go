package console

import (
	"github.com/michael-reichenauer/gmc/api"
	"github.com/michael-reichenauer/gmc/utils/cui"
)

type DetailsView struct {
	cui.View
	vm *detailsVM
}

func NewDetailsView(uiHandler cui.UI) *DetailsView {
	h := &DetailsView{vm: NewDetailsVM()}
	h.View = uiHandler.NewViewFromTextFunc(h.viewData)
	h.View.Properties().Name = "DetailsView"
	return h
}

func (h *DetailsView) viewData(viewPage cui.ViewPage) string {
	details, err := h.vm.getCommitDetails(viewPage)
	if err != nil {
		return ""
	}
	return details
}

func (h *DetailsView) SetCurrent(commit api.Commit) {
	h.vm.setCurrentCommit(commit)
	h.NotifyChanged()
}
