package console

import (
	"fmt"

	"github.com/michael-reichenauer/gmc/utils/cui"
	"github.com/michael-reichenauer/gmc/utils/git"
)

func ShowAboutDlg(ui cui.UI) {
	gitVersion := git.Version()
	ui.ShowMessageBox("About gmc", fmt.Sprintf("gmc: %s\ngit: %s",
		ui.Version(), gitVersion))
}
