package repoview

import (
	"fmt"
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
		return ui.ViewData{Text: ui.Red(fmt.Sprintf("Error: %v", err)), FirstLine: 0, Lines: 1, TotalLines: 1}
	}
	return ui.ViewData{Text: details.Text, FirstLine: 0, Lines: 1, TotalLines: 4}
}

func (h *DetailsView) SetCurrent(selectedIndex int) {
	h.selectedIndex = selectedIndex
	h.NotifyChanged()
}
