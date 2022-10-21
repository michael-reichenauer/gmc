package console

import (
	"fmt"

	"github.com/jroimartin/gocui"
	"github.com/michael-reichenauer/gmc/api"
	"github.com/michael-reichenauer/gmc/utils/cui"
	"github.com/michael-reichenauer/gmc/utils/log"
)

type DetailsView struct {
	cui.View
	vm       *detailsVM
	repoView *RepoView
}

func NewDetailsView(ui cui.UI, repoView *RepoView) *DetailsView {
	t := &DetailsView{repoView: repoView}

	t.View = ui.NewViewFromTextFunc(t.viewData)
	t.View.Properties().Name = "DetailsView"
	t.View.Properties().Title = "Commit Details"

	t.View.Properties().HideCurrentLineMarker = true
	t.View.Properties().HideHorizontalScrollbar = true
	t.View.Properties().HasFrame = true
	t.View.SetKey(gocui.KeyEnter, t.onClose)
	t.View.SetKey(gocui.KeyEsc, t.onClose)
	t.View.SetKey(gocui.KeyCtrlC, t.onClose)
	t.View.SetKey(gocui.KeyTab, t.onKeyTab)
	t.View.SetKey('d', t.repoView.vm.showSelectedCommitDiff)
	t.View.SetKey('D', t.repoView.vm.showSelectedCommitDiff)
	t.View.SetKey(gocui.KeyCtrlD, t.repoView.vm.showSelectedCommitDiff)

	t.vm = NewDetailsVM(t.View)
	return t
}

func (t *DetailsView) SetCurrentView() {
	t.View.Properties().Title = "* Commit Details *"
	t.View.NotifyChanged()
	t.View.SetCurrentView()
}

func (t *DetailsView) SetCurrentLine(line int, repo api.Repo, repoId string, api api.Api) {
	log.Infof("line %d %#v", line, t.vm)
	t.vm.setCurrentLine(line, repo, repoId, api)
}

func (h *DetailsView) viewData(viewPage cui.ViewPage) string {
	details, err := h.vm.getCommitDetails(viewPage)
	if err != nil {
		return cui.Red(fmt.Sprintf("Failed to get commit details:\n%s", err))
	}
	return details
}

func (t *DetailsView) onClose() {
	t.repoView.hideCommitDetails()
}

func (t *DetailsView) onKeyTab() {
	t.View.Properties().Title = "Commit Details"
	t.View.NotifyChanged()
	t.repoView.SetCurrentView()
	t.repoView.NotifyChanged()
}
