package main

import (
	"flag"
	"fmt"
	"github.com/mattn/go-isatty"
	"github.com/michael-reichenauer/gmc/common/config"
	"github.com/michael-reichenauer/gmc/installation"
	"github.com/michael-reichenauer/gmc/program"
	"github.com/michael-reichenauer/gmc/utils"
	"github.com/michael-reichenauer/gmc/utils/git"
	"github.com/michael-reichenauer/gmc/utils/log"
	"github.com/michael-reichenauer/gmc/utils/log/logger"
	"github.com/michael-reichenauer/gmc/utils/ui"
	"io/ioutil"
	stdlog "log"
	"os"
	"os/exec"
	"runtime"
)

const (
	version = "0.22"
)

var (
	repoPathFlag    = flag.String("d", "", "specify working folder")
	showVersionFlag = flag.Bool("version", false, "print gmc version")
)

func main() {
	flag.Parse()
	if *showVersionFlag {
		fmt.Printf("%s", version)
		return
	}

	if isDebugConsole() {
		// Seems program is run within a debugger console, which the termbox does not support
		// So a new external process has been started and this instance ends
		return
	}
	// Disable standard logging since some modules log to stderr, which conflicts with console ui
	stdlog.SetOutput(ioutil.Discard)
	// Set default http client proxy to the system proxy (used by e.g. telemetry)
	utils.SetDefaultHTTPProxy()
	logger.StdTelemetry.Enable(version)

	log.Eventf("program-start", "Starting gmc %s ...", version)

	configService := config.NewConfig(version, *repoPathFlag)
	configService.SetState(func(s *config.State) {
		s.InstalledVersion = version
	})

	logProgramInfo(configService)

	autoUpdate := installation.NewAutoUpdate(configService, version)
	autoUpdate.Start()

	uiHandler := ui.NewUI()
	uiHandler.Run(func() {
		mainWindow := program.NewMainWindow(uiHandler, configService)
		uiHandler.OnResizeWindow = mainWindow.OnResizeWindow
		mainWindow.Show()
	})

	log.Event("program-stop")
	logger.StdTelemetry.Close()
}

func isDebugConsole() bool {
	// termbox requires a "real" terminal, so check if current terminal is ok (check only on windows)
	if isatty.IsTerminal(os.Stdout.Fd()) || isatty.IsCygwinTerminal(os.Stdout.Fd()) ||
		runtime.GOOS != "windows" {
		return false
	}

	// Seems to running in a debug terminal like e.g. in goland, restart as an external command
	args := []string{"/C", "start"}
	args = append(args, os.Args...)
	cmd := exec.Command("cmd", args...)
	_ = cmd.Start()
	_ = cmd.Wait()
	return true
}

func logProgramInfo(configService *config.Service) {
	log.Infof("Version: %s", configService.ProgramVersion)
	log.Infof("Binary path: %q", utils.BinPath())
	log.Infof("Folder path: %s", configService.FolderPath)
	log.Infof("Git version: %s", git.GitVersion())
	log.Infof("Http proxy: %s", utils.GetHTTPProxyURL())
}
