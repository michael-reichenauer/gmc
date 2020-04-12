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
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/exec"
	"runtime"
)

const (
	version = "0.26"
)

var (
	workingFolderFlag = flag.String("d", "", "specify working folder")
	showVersionFlag   = flag.Bool("version", false, "print gmc version")
	pauseFlag         = flag.Bool("pause", false, "pause until user click enter")
)

func main() {
	flag.Parse()
	if *showVersionFlag {
		fmt.Printf("%s", version)
		return
	}

	if isDebugConsole() {
		// Seems program is run within a debugger console, which the termbox does not support
		// So a new external process is started and this instance ends
		startAsExternalProcess()
		return
	}

	go func() {
		log.Infof("prof on port 6060 %v", http.ListenAndServe("localhost:6060", nil))
	}()

	if *pauseFlag {
		// The process was started with 'pause' flag, e.g. from Goland,
		// wait until enter is clicked, this can be used for attaching a debugger
		fmt.Printf("Click 'enter' to proceed ...\n")
		utils.ReadLine()
	}

	// Disable standard logging since some modules log to stderr, which conflicts with console ui
	stdlog.SetOutput(ioutil.Discard)
	// Set default http client proxy to the system proxy (used by e.g. telemetry)
	utils.SetDefaultHTTPProxy()
	logger.StdTelemetry.Enable(version)
	defer logger.StdTelemetry.Close()

	log.Eventf("program-start", "Starting gmc %s ...", version)
	defer log.Event("program-stop")

	configService := config.NewConfig(version, *workingFolderFlag)
	configService.SetState(func(s *config.State) {
		s.InstalledVersion = version
	})

	logProgramInfo(configService)

	autoUpdate := installation.NewAutoUpdate(configService, version)
	autoUpdate.Start()

	uiHandler := ui.NewUI()
	uiHandler.Run(func() {
		mainWindow := program.NewMainWindow(uiHandler, configService)
		mainWindow.Show()
	})
}

func isDebugConsole() bool {
	// termbox requires a "real" terminal. The GoLand console is not supported.
	// so check if current terminal is ok (check only on windows)
	if isatty.IsTerminal(os.Stdout.Fd()) || isatty.IsCygwinTerminal(os.Stdout.Fd()) ||
		runtime.GOOS != "windows" {
		return false
	}
	return true
}

func startAsExternalProcess() {
	args := []string{"/C", "start"}
	args = append(args, os.Args...)
	cmd := exec.Command("cmd", args...)
	_ = cmd.Start()
	_ = cmd.Wait()
}

func logProgramInfo(configService *config.Service) {
	log.Infof("Version: %s", configService.ProgramVersion)
	log.Infof("Binary path: %q", utils.BinPath())
	log.Infof("OS: %q", runtime.GOOS)
	log.Infof("Arch: %q", runtime.GOARCH)
	log.Infof("Go version: %q", runtime.Version())
	log.Infof("Folder path: %q", configService.FolderPath)
	log.Infof("Working Folder: %q", utils.CurrentDir())
	log.Infof("Git version: %q", git.Version())
	log.Infof("Http proxy: %q", utils.GetHTTPProxyURL())
}
