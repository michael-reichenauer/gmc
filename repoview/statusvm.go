package repoview

import (
	"fmt"
	"github.com/michael-reichenauer/gmc/repoview/model"
	"github.com/michael-reichenauer/gmc/utils"
	"github.com/michael-reichenauer/gmc/utils/ui"
)

type statusVM struct {
	model *model.Model
}

func newStatusVM(model *model.Model) *statusVM {
	return &statusVM{model: model}
}

func (h *statusVM) GetStatus(width int) (string, error) {
	status := h.model.GetStatus()
	if status.AllChanges == 0 {
		return "", nil
	}
	gw := status.GraphWidth + 4

	width = width - gw
	if width < 0 {
		return "", nil
	}
	pad := ""
	for i := 0; i < gw; i++ {
		pad = pad + " "
	}

	statusText := utils.Text(fmt.Sprintf("%d uncommited chages", status.AllChanges), width)
	return pad + ui.YellowDk(statusText), nil
}
