package main

import (
	"gmc/repoview"
	"gmc/utils"
	"gmc/utils/ui"
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
