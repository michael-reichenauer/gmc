package main

import (
	"flag"
	"github.com/mattn/go-isatty"
	"github.com/michael-reichenauer/gmc/repoview"
	"github.com/michael-reichenauer/gmc/utils"
	"github.com/michael-reichenauer/gmc/utils/ui"
	"os"
	"os/exec"
	"runtime"
)

const version = "0.3"

var (
	repoPath = flag.String("d", "", "Specify directory")
)

func main() {
	flag.Parse()
	if *repoPath == "" {
		// No specified repo path, use current dir
		*repoPath = utils.CurrentDir()
	}

	if !isatty.IsTerminal(os.Stdout.Fd()) && !isatty.IsCygwinTerminal(os.Stdout.Fd()) && runtime.GOOS == "windows" {
		// Seems to be not running in a terminal like e.g. in goland,
		// termbox requires a terminal, so lets restart as an external command on windows
		var args []string
		args = append(args, "/C")
		args = append(args, "start")
		for _, arg := range os.Args {
			args = append(args, arg)
		}
		cmd := exec.Command("cmd", args...)
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
