package console

import (
	"github.com/michael-reichenauer/gmc/doc"
	"github.com/michael-reichenauer/gmc/utils/cui"
)

func ShowHelpDlg(ui cui.UI) {
	ui.ShowMessageBox("Help", doc.HelpFile)
}
