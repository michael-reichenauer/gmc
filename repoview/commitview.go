package repoview

import (
	"github.com/jroimartin/gocui"
	"github.com/michael-reichenauer/gmc/utils/ui"
)

type CommitView struct {
	ui.View
	vm          *commitVM
	ui          *ui.UI
	mainService mainService
}

func NewCommitView(ui *ui.UI, mainService mainService, branchName string) *CommitView {
	h := &CommitView{ui: ui, vm: NewCommitVM(), mainService: mainService}
	h.View = ui.NewViewFromTextFunc(h.viewData)
	h.View.Properties().Name = "CommitView"
	h.View.Properties().Title = "Commit on: " + branchName
	h.View.Properties().IsEditable = true
	h.View.Properties().OnLoad = h.onLoad
	return h
}

func (h *CommitView) viewData(viewPage ui.ViewPage) string {
	details, err := h.vm.getCommitDetails(viewPage)
	if err != nil {
		return ""
	}
	return details
}

func (h *CommitView) onLoad() {
	h.SetKey(gocui.KeyEsc, gocui.ModNone, h.mainService.HideCommit)
}
