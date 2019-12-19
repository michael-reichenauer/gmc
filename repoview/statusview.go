package repoview

import (
	"github.com/michael-reichenauer/gmc/utils/log"
	"github.com/michael-reichenauer/gmc/utils/ui"
)

type StatusViewHandler struct {
	ui.View
	vm *statusVM
}

func NewStatusView(uiHandler *ui.UI) *StatusViewHandler {
	h := &StatusViewHandler{
		vm: newStatusVM(),
	}
	h.View = uiHandler.NewView(h.viewData)
	h.Properties().OnLoad = h.onLoad
	return h
}

func (h *StatusViewHandler) viewData(viewPort ui.ViewPort) ui.ViewData {
	//repoPage, err := h.vm.GetRepoPage(viewPort.Width, viewPort.First, viewPort.Last, viewPort.Current)
	//if err != nil {
	//	return ui.viewData{Text: ui.Red(fmt.Sprintf("Error: %v", err)), MaxLines: 1}
	//}
	log.Infof("status")
	return ui.ViewData{
		Text:     "status text",
		MaxLines: 1,
		First:    0,
		Last:     1,
		Current:  1,
	}
}

func (h *StatusViewHandler) onLoad() {
	//h.vm.Load()

	h.NotifyChanged()
}
