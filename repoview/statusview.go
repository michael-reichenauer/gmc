package repoview

import (
	"fmt"
	"github.com/michael-reichenauer/gmc/repoview/model"
	"github.com/michael-reichenauer/gmc/utils/log"
	"github.com/michael-reichenauer/gmc/utils/ui"
)

type StatusViewHandler struct {
	ui.View
	vm *statusVM
}

func newStatusView(uiHandler *ui.UI, model *model.Model) *StatusViewHandler {
	h := &StatusViewHandler{
		vm: newStatusVM(model),
	}
	h.View = uiHandler.NewView(h.viewData)
	return h
}

func (h *StatusViewHandler) viewData(viewPort ui.ViewPort) ui.ViewData {
	log.Infof("Get status")
	status, err := h.vm.GetStatus(viewPort.Width)
	if err != nil {
		return ui.ViewData{Text: ui.Red(fmt.Sprintf("Error: %v", err)), MaxLines: 1}
	}
	log.Infof("status: %s", status)
	return ui.ViewData{Text: status, MaxLines: 1, First: 0, Last: 1, Current: 1}
}
