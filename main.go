package main

import (
	"github.com/michael-reichenauer/gmc/repoview"
	"github.com/michael-reichenauer/gmc/utils"
	"github.com/michael-reichenauer/gmc/utils/ui"
)

func main() {
	//repoPath := `C:\Work Files\GitMind`
	repoPath := utils.CurrentDir()

	uiHandler := ui.NewUI()
	uiHandler.Run(func() {
		repoView := repoview.New(uiHandler, repoPath)
		_ = uiHandler.Show(repoView)
	})
}
