package program

import (
	"flag"
	"github.com/mattn/go-isatty"
	"github.com/michael-reichenauer/gmc/common/config"
	"github.com/michael-reichenauer/gmc/repoview"
	"github.com/michael-reichenauer/gmc/utils"
	"github.com/michael-reichenauer/gmc/utils/gitlib"
	"github.com/michael-reichenauer/gmc/utils/log"
	"github.com/michael-reichenauer/gmc/utils/ui"
	"os"
	"os/exec"
	"runtime"
)

var (
	repoPath = flag.String("d", "", "Specify directory")
)

func Main(version string) {
	flag.Parse()
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
	log.Infof("Starting gmc version: %s ...", version)
	if *repoPath == "" {
		// No specified repo path, use current dir
		*repoPath = utils.CurrentDir()
	}

	path, err := gitlib.WorkingFolderRoot(*repoPath)
	if err != nil {
		panic(log.Error(err))
	}
	configService := config.NewConfig()
	configService.Load()

	uiHandler := ui.NewUI()
	uiHandler.Run(func() {
		mainWindow := repoview.NewMainWindow(uiHandler, configService, path)
		uiHandler.OnResizeWindow = mainWindow.OnResizeWindow
		mainWindow.Show()
	})
}
