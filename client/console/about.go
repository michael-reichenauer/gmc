package console

import (
	"fmt"

	"github.com/michael-reichenauer/gmc/utils/cui"
)

func ShowAboutDlg(ui cui.UI) {
	ui.ShowMessageBox("About gmc", fmt.Sprintf("Version: %s", ui.Version()))
}
