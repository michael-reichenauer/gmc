package program

import (
	"flag"
	"fmt"
	"github.com/mattn/go-isatty"
	"github.com/michael-reichenauer/gmc/common/config"
	"github.com/michael-reichenauer/gmc/installation"
	"github.com/michael-reichenauer/gmc/utils"
	"github.com/michael-reichenauer/gmc/utils/log"
	"os"
	"os/exec"
	"runtime"
)

var (
	repoPathFlag = flag.String("d", "", "specify working folder")
	versionFlag  = flag.Bool("version", false, "print gmc version")
)

func Main(version string) {
	flag.Parse()
	if *versionFlag {
		fmt.Printf("%s", version)
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
	log.Infof("Starting gmc version: %s %q ...", version, utils.BinPath())
	configService := config.NewConfig()
	au := installation.NewAutoUpdate(configService, version)
	au.Start()

	// if *repoPath == "" {
	// 	// No specified repo path, use current dir
	// 	*repoPath = utils.CurrentDir()
	// }
	//
	// path, err := gitlib.WorkingFolderRoot(*repoPath)
	// if err != nil {
	// 	panic(log.Error(err))
	// }
	//
	// uiHandler := ui.NewUI()
	// uiHandler.Run(func() {
	// 	mainWindow := repoview.NewMainWindow(uiHandler, configService, path)
	// 	uiHandler.OnResizeWindow = mainWindow.OnResizeWindow
	// 	mainWindow.Show()
	//})
}
