package main

import (
	"flag"
	"github.com/michael-reichenauer/gmc/repoview"
	"github.com/michael-reichenauer/gmc/utils"
	"github.com/michael-reichenauer/gmc/utils/ui"
	"os"
	"os/exec"
)

var (
	repoPath = flag.String("d", "", "Specify directory")
	external = flag.Bool("external", false, "Specify directory")
)

var version string

func main() {
	flag.Parse()
	if *repoPath == "" {
		*repoPath = utils.CurrentDir()
	}
	//*repoPath = `C:\Work Files\GitMind`
	if *external {
		var arg []string
		arg = append(arg, "/C")
		arg = append(arg, "start")
		for i := 0; i < len(os.Args); i++ {
			if os.Args[i] != "-external" {
				arg = append(arg, os.Args[i])
			}
		}
		cmd := exec.Command("cmd", arg...)
		_ = cmd.Start()
		_ = cmd.Wait()
		return
	}
	uiHandler := ui.NewUI()
	uiHandler.Run(func() {
		repoView := repoview.New(uiHandler, *repoPath)
		_ = uiHandler.Show(repoView)
	})
}
