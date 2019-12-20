package repoview

import (
	"github.com/michael-reichenauer/gmc/repoview/model"
	"github.com/michael-reichenauer/gmc/utils/log"
	"github.com/michael-reichenauer/gmc/utils/ui"
)

type DetailsView struct {
	ui.View
	vm *detailsVM
}

func newDetailsView(uiHandler *ui.UI, model *model.Model) *DetailsView {
	h := &DetailsView{
		vm: newDetailsVM(model),
	}
	h.View = uiHandler.NewView(h.viewData)
	return h
}

func (h *DetailsView) viewData(viewPort ui.ViewPort) ui.ViewData {
	log.Infof("Get details")
	return ui.ViewData{Text: "commit details", MaxLines: 1, First: 0, Last: 1, Current: 1}
}
