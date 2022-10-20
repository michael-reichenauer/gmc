package console

import (
	"github.com/jroimartin/gocui"
	"github.com/michael-reichenauer/gmc/api"
	"github.com/michael-reichenauer/gmc/utils/cui"
	"github.com/michael-reichenauer/gmc/utils/log"
)

type DetailsView struct {
	cui.View
	vm *detailsVM
}

func NewDetailsView(uiHandler cui.UI) *DetailsView {
	h := &DetailsView{}
	h.vm = NewDetailsVM(h)
	h.View = uiHandler.NewViewFromTextFunc(h.viewData)
	h.View.Properties().Name = "DetailsView"
	h.View.Properties().Title = "Commit Details"

	h.View.Properties().HideCurrentLineMarker = true
	h.View.Properties().HideHorizontalScrollbar = true
	h.View.Properties().HasFrame = true
	h.View.SetKey(gocui.KeyEnter, h.View.Close)

	h.View.SetKey(gocui.KeyEsc, h.View.Close)
	h.View.SetKey(gocui.KeyCtrlC, h.View.Close)

	return h
}

func (h *DetailsView) viewData(viewPage cui.ViewPage) string {
	details, err := h.vm.getCommitDetails(viewPage)
	if err != nil {
		return ""
	}
	return details
}

func (h *DetailsView) SetCurrent(line int, repo api.Repo, repoId string, api api.Api) {
	log.Infof("line %d %#v", line, h.vm)
	h.vm.setCurrentLine(line, repo, repoId, api)
}
