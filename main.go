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
		args := []string{"/C", "start"}
		args = append(args, os.Args...)
		cmd := exec.Command("cmd", args...)
		_ = cmd.Start()
		_ = cmd.Wait()
		return
	}

	uiHandler := ui.NewUI()
	uiHandler.Run(func() {
		mainWindow := repoview.NewMainWindow(uiHandler, *repoPath)
		uiHandler.OnResizeWindow = mainWindow.OnResizeWindow
		mainWindow.Show()
	})
}
