package repoview

import (
	"github.com/michael-reichenauer/gmc/repoview/model"
)

type detailsVM struct {
	model *model.Model
}

func newDetailsVM(model *model.Model) *detailsVM {
	return &detailsVM{model: model}
}

//func (h *statusVM) GetStatus(width int) (string, error) {
//	status := h.model.GetStatus()
//	if status.AllChanges == 0 {
//		return "", nil
//	}
//	gw := status.GraphWidth + 4
//
//	width = width - gw
//	if width < 0 {
//		return "", nil
//	}
//	pad := ""
//	//for i := 0; i < gw; i++ {
//	//	pad = pad + " "
//	//}
//
//	statusText := utils.Text(fmt.Sprintf(" %d uncommited changes", status.AllChanges), width)
//	return pad + ui.YellowDk(statusText), nil
//}
